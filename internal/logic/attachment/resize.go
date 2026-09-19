package attachment
import (
	"bytes"
	"container/list"
	"encoding/binary"
	"hash/fnv"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"sync"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
)

// thumbnailMaxWidth 缩略图最大宽度（卡片网格 2~4 列布局，约 300px 显示宽度 * 2 倍屏）
const thumbnailMaxWidth = 360

// resizeCacheWidths 允许的缩放宽度档位（防滥用 + 利于缓存命中）
var resizeCacheWidths = map[int]bool{thumbnailMaxWidth: true}

// cacheEntry 缩放结果缓存条目
type cacheEntry struct {
	key   thumbKey
	bytes []byte
}

// thumbKey 缓存键：附件 id + 原始内容指纹（FNV-1a 64 位，等价内容同键）
type thumbKey struct {
	id   int64
	fp   [8]byte
}

func newThumbKey(id int64, content []byte) thumbKey {
	h := fnv.New64a()
	h.Write(content) // hash.Write never errors
	var fp [8]byte
	binary.BigEndian.PutUint64(fp[:], h.Sum64())
	return thumbKey{id: id, fp: fp}
}

// thumbCache 简单 LRU：按 (id, 内容指针) 缓存缩放结果，避免同一 blob 反复解码缩放
type thumbCache struct {
	mu    sync.Mutex
	cap   int
	ll    *list.List // front = newest
	items map[thumbKey]*list.Element
}

func newThumbCache(cap int) *thumbCache {
	return &thumbCache{cap: cap, ll: list.New(), items: make(map[thumbKey]*list.Element, cap)}
}

func (c *thumbCache) get(k thumbKey) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[k]; ok {
		c.ll.MoveToFront(el)
		return el.Value.(*cacheEntry).bytes, true
	}
	return nil, false
}

func (c *thumbCache) put(k thumbKey, b []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[k]; ok {
		c.ll.MoveToFront(el)
		el.Value.(*cacheEntry).bytes = b
		return
	}
	el := c.ll.PushFront(&cacheEntry{key: k, bytes: b})
	c.items[k] = el
	for c.ll.Len() > c.cap {
		back := c.ll.Back()
		if back == nil {
			break
		}
		delete(c.items, back.Value.(*cacheEntry).key)
	}
}

var thumbs = newThumbCache(64)

// webpQuality webp 有损编码质量（卡片含 AI 生成文字，取 85 保证文字清晰）
const webpQuality = 85

// ResizeBytes 将图片二进制按目标宽度等比缩小（不放大），输出 webp 编码。
// 仅对栅格图（jpeg/png/webp/gif）生效；其他类型原样返回（resized=false）。
func ResizeBytes(content []byte, contentType string, targetWidth int) (out []byte, mime string, resized bool) {
	if !resizeCacheWidths[targetWidth] || targetWidth <= 0 {
		return content, contentType, false
	}

	// 缓存命中：同一 blob 的同宽度结果直接复用
	key := newThumbKey(0, content)
	if cached, ok := thumbs.get(key); ok {
		return cached, "image/webp", true
	}

	// 解码：按实际 MIME 分派（嗅探兜底）
	src, err := decodeImage(content, contentType)
	if err != nil {
		return content, contentType, false // 解码失败：原样返回，不阻塞请求
	}
	b := src.Bounds()
	if b.Dx() <= targetWidth {
		return content, contentType, false // 原图不比目标宽：无需缩放
	}
	targetH := b.Dy() * targetWidth / b.Dx()

	dst := image.NewNRGBA(image.Rect(0, 0, targetWidth, targetH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Src, nil)

	var buf bytes.Buffer
	if err := webp.Encode(&buf, dst, webp.Options{Quality: webpQuality}); err != nil {
		return content, contentType, false // 编码失败：原样返回
	}
	out = buf.Bytes()
	thumbs.put(key, out)
	return out, "image/webp", true
}

// ToWebpBytes 将栅格图二进制转为 webp（等尺寸重编码，不缩放）。
// 已是 webp、svg/矢量或解码失败时原样返回（converted=false）。
func ToWebpBytes(content []byte, contentType string) (out []byte, mime string, converted bool) {
	switch contentType {
	case "image/webp", "image/svg+xml", "image/gif":
		return content, contentType, false // 已最优/矢量/动图：不动
	}
	src, err := decodeImage(content, contentType)
	if err != nil {
		return content, contentType, false
	}
	// 统一过 NRGBA（webp 编码器内部同样转换，显式化避免双份拷贝歧义）
	dst := image.NewNRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	var buf bytes.Buffer
	if err := webp.Encode(&buf, dst, webp.Options{Quality: webpQuality}); err != nil {
		return content, contentType, false
	}
	// 体积不减反增（已高度压缩的小 JPEG）：保留原格式
	if buf.Len() >= len(content) {
		return content, contentType, false
	}
	return buf.Bytes(), "image/webp", true
}

func decodeImage(content []byte, contentType string) (image.Image, error) {
	r := bytes.NewReader(content)
	switch contentType {
	case "image/png":
		return png.Decode(r)
	case "image/jpeg":
		return jpeg.Decode(r)
	case "image/webp":
		return webp.Decode(r)
	default:
		if img, _, err := image.Decode(bytes.NewReader(content)); err == nil {
			return img, nil
		}
		r2 := bytes.NewReader(content)
		return webp.Decode(io.MultiReader(r2))
	}
}
