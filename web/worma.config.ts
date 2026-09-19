import { defineConfig } from 'wormajs';
import type { Api, ApiDoc, TemplateConfigResult, TemplateData } from 'wormajs';
import type * as HandlebarsModule from 'handlebars';

/**
 * operationId 拆分插件：
 * `skill_getList` → tag=skill（资源命名空间） + name=getList（方法名）
 *
 * 后端约定见根目录 AGENTS.md：operationId 必须为 `资源_动作`（下划线分隔）。
 * 未带分隔符的 operationId 保留原 OpenAPI tag 分组。
 */
function splitOperationId() {
  const split = (api: Api): void => {
    const match = /^([A-Za-z][\w-]*)[_-](.+)$/.exec(api.name);
    if (!match) return;
    api.tag = match[1].replace(/-+(.)/g, (_, c: string) => c.toUpperCase());
    api.name = match[2].replace(/[_-]+(.)/g, (_, c: string) => c.toUpperCase());
  };
  return {
    name: 'split-operation-id',
    beforeCodeGenerate({ data }: { data: TemplateData }): void {
      data.allApis.forEach(split);
      const groups = new Map<string, ApiDoc>();
      for (const api of data.allApis) {
        const doc = groups.get(api.tag) ?? { tag: api.tag, apis: [] };
        doc.apis.push(api);
        groups.set(api.tag, doc);
      }
      data.tagedApis = [...groups.values()];
    },
  };
}

/** JS/TS 保留字，作为导出函数名时需走 `export { fn as name }` 别名 */
const RESERVED_WORDS = new Set([
  'await', 'break', 'case', 'catch', 'class', 'const', 'continue', 'debugger', 'default', 'delete', 'do', 'else',
  'enum', 'export', 'extends', 'false', 'finally', 'for', 'function', 'if', 'implements', 'import', 'in',
  'instanceof', 'interface', 'let', 'new', 'null', 'package', 'private', 'protected', 'public', 'return', 'static',
  'super', 'switch', 'this', 'throw', 'true', 'try', 'typeof', 'var', 'void', 'while', 'with', 'yield',
]);

function nuxtRequesterTemplate() {
  return {
    name: 'nuxt-requester-template',
    getTemplate: (): TemplateConfigResult => ({ path: './worma/templates/nuxt' }),
    onHandlebarsCreated({ hbs }: { hbs: typeof HandlebarsModule }): void {
      hbs.registerHelper('isReserved', (value: unknown) => RESERVED_WORDS.has(String(value)));
    },
  };
}

export default defineConfig({
  generator: [
    {
      input: 'http://127.0.0.1:8000/api.json',
      output: 'app/api',
      plugins: [splitOperationId(), nuxtRequesterTemplate()],
    },
  ],
});
