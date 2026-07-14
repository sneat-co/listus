import nx from '@nx/eslint-plugin';

export default [
  ...nx.configs['flat/base'],
  ...nx.configs['flat/typescript'],
  ...nx.configs['flat/javascript'],
  {
    ignores: [
      '**/dist',
      '**/vite.config.*.timestamp*',
      '**/vitest.config.*.timestamp*',
    ],
  },
  {
    files: ['**/*.ts', '**/*.tsx', '**/*.js', '**/*.jsx'],
    rules: {
      '@nx/enforce-module-boundaries': [
        'error',
        {
          enforceBuildableLibDependency: true,
          allow: ['^.*/eslint(\\.base)?\\.config\\.[cm]?[jt]s$'],
          depConstraints: [
            {
              sourceTag: 'layer:contract',
              onlyDependOnLibsWithTags: [
                'layer:contract',
                'scope:foundation',
              ],
            },
            {
              sourceTag: 'layer:ui',
              onlyDependOnLibsWithTags: [
                'layer:contract',
                'layer:ui',
                'scope:foundation',
              ],
            },
            {
              sourceTag: 'layer:runtime',
              onlyDependOnLibsWithTags: [
                'layer:contract',
                'layer:ui',
                'layer:runtime',
                'scope:foundation',
              ],
            },
            {
              sourceTag: 'layer:app',
              onlyDependOnLibsWithTags: [
                'layer:contract',
                'layer:ui',
                'layer:runtime',
                'scope:foundation',
              ],
            },
            {
              sourceTag: 'layer:e2e',
              onlyDependOnLibsWithTags: ['layer:app'],
            },
          ],
        },
      ],
    },
  },
  {
    files: [
      '**/*.ts',
      '**/*.tsx',
      '**/*.cts',
      '**/*.mts',
      '**/*.js',
      '**/*.jsx',
      '**/*.cjs',
      '**/*.mjs',
    ],
    // Override or add rules here
    rules: {},
  },
];
