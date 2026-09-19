package main

import (
	_ "cookbook/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"cookbook/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
