import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import Nimma from 'nimma';
import { jsonPathPlus } from 'nimma/fallbacks';

const directory = path.dirname(fileURLToPath(import.meta.url));
const corpusPath = path.resolve(directory, '..', 'corpus.json');
const corpus = JSON.parse(await fs.readFile(corpusPath, 'utf8'));
const officialSelectorsPath = path.resolve(directory, '..', 'official-selectors.json');
const officialInputPath = path.resolve(directory, '..', 'official-input.json');
const officialParityPath = path.resolve(directory, '..', 'official-parity.json');

function normalizedPath(root, segments) {
  let result = '$';
  let current = root;
  for (const segment of segments) {
    if (Array.isArray(current)) {
      result += `[${segment}]`;
    } else {
      result += `['${String(segment).replaceAll('\\', '\\\\').replaceAll("'", "\\'")}']`;
    }
    current = current?.[segment];
  }
  return result;
}

function selectedEngine(expression) {
  try {
    new Nimma([expression], { fallback: null, unsafe: false, output: 'auto' });
    return 'nimma';
  } catch {
    return 'jsonpath-plus';
  }
}

function executeCase(testCase) {
  const paths = [];
  const query = new Nimma([testCase.expression], {
    fallback: jsonPathPlus,
    unsafe: false,
    output: 'auto',
    customShorthands: {},
  });
  query.query(testCase.input, {
    [testCase.expression]: scope => paths.push(normalizedPath(testCase.input, scope.path)),
  });
  testCase.engine = selectedEngine(testCase.expression);
  testCase.expectedPaths = paths.sort();
  return testCase;
}

for (const testCase of corpus.cases) {
  executeCase(testCase);
}

corpus.baseline.generatedBy = 'npm ci && npm run generate';
await fs.writeFile(corpusPath, `${JSON.stringify(corpus, null, 2)}\n`);

const officialSelectors = JSON.parse(await fs.readFile(officialSelectorsPath, 'utf8'));
const officialInput = JSON.parse(await fs.readFile(officialInputPath, 'utf8'));
const officialParity = {
  baseline: {
    spectralCommit: corpus.baseline.spectralCommit,
    spectralCoreVersion: corpus.baseline.spectralCoreVersion,
    nimmaVersion: corpus.baseline.nimmaVersion,
    jsonPathPlusVersion: corpus.baseline.jsonPathPlusVersion,
    unsafe: false,
    generatedBy: 'npm ci && npm run collect-official && npm run generate',
    sourceRepository: officialSelectors.repository,
    sourceLicense: officialSelectors.license,
  },
  input: officialInput,
  cases: officialSelectors.selectors.map(expression => executeCase({ expression, input: officialInput })).map(testCase => ({
    expression: testCase.expression,
    engine: testCase.engine,
    expectedPaths: testCase.expectedPaths,
  })),
};
await fs.writeFile(officialParityPath, `${JSON.stringify(officialParity, null, 2)}\n`);
