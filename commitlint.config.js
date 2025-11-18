// commitlint.config.js
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // Increase max body line length (e.g., to 200). Use 0 to disable the rule entirely.
    'body-max-line-length': [2, 'always', 300],
  },
};
