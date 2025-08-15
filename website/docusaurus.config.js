// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'github-comment',
  tagline: 'CLI to create and hide GitHub comments',
  url: 'https://suzuki-shunsuke.github.io',
  baseUrl: '/github-comment/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'suzuki-shunsuke', // Usually your GitHub org/user name.
  projectName: 'github-comment', // Usually your repo name.

  presets: [
    [
      '@docusaurus/preset-classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/suzuki-shunsuke/github-comment-docs/edit/main',
          routeBasePath: '/',
        },
        pages: false,
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'github-comment',
        items: [
          {
            href: 'https://github.com/suzuki-shunsuke/github-comment',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Community',
            items: [],
          },
          {
            title: 'More',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/suzuki-shunsuke/github-comment',
              },
            ],
          },
        ],
        copyright: `Copyright © 2020 Shunsuke Suzuki. Built with Docusaurus.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      algolia: {
        appId: 'QUD6958438',
        // Public API key: it is safe to commit it
        apiKey: '1f3fd0dfcc22bc13977eefc554b422eb',
        indexName: 'github-comment',
        searchParameters: {},
      },
    }),
};

module.exports = config;
