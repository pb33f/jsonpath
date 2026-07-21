import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const commit = 'f08df7bb4c62015391cc2b0d51cdaecf30ac109c';
const repository = 'https://github.com/stoplightio/spectral';
const license = 'Apache-2.0';
const files = [
  'packages/rulesets/src/oas/index.ts',
  'packages/rulesets/src/asyncapi/index.ts',
];
const base = `https://raw.githubusercontent.com/stoplightio/spectral/${commit}/`;
const selectors = new Set();

function decodeString(quote, body) {
  if (quote === '"') {
    return JSON.parse(`"${body}"`);
  }
  return body
    .replaceAll(`\\'`, `'`)
    .replaceAll('\\"', '"')
    .replaceAll('\\\\', '\\');
}

function collectStrings(source) {
  const values = [];
  for (let index = 0; index < source.length; ) {
    if (source.startsWith('//', index)) {
      index = source.indexOf('\n', index + 2);
      if (index === -1) break;
      continue;
    }
    if (source.startsWith('/*', index)) {
      const end = source.indexOf('*/', index + 2);
      index = end === -1 ? source.length : end + 2;
      continue;
    }
    const quote = source[index];
    if (quote !== '"' && quote !== "'" && quote !== '`') {
      index++;
      continue;
    }
    const start = ++index;
    let body = '';
    while (index < source.length && source[index] !== quote) {
      if (source[index] === '\\' && index + 1 < source.length) {
        body += source[index] + source[index + 1];
        index += 2;
      } else {
        body += source[index++];
      }
    }
    if (quote !== '`') {
      values.push(decodeString(quote, body));
    }
    index = index < source.length ? index + 1 : source.length;
  }
  return values;
}

for (const file of files) {
  const response = await fetch(base + file);
  if (!response.ok) {
    throw new Error(`failed to fetch ${file}: ${response.status} ${response.statusText}`);
  }
  const source = await response.text();
  for (const value of collectStrings(source)) {
    if (value.startsWith('$') && !value.startsWith('${')) {
      selectors.add(value);
    }
  }
}

const directory = path.dirname(fileURLToPath(import.meta.url));
const outputPath = path.resolve(directory, '..', 'official-selectors.json');
await fs.writeFile(outputPath, `${JSON.stringify({ repository, license, commit, files, selectors: [...selectors].sort() }, null, 2)}\n`);
