module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true
  },
  extends: ["eslint:recommended", "plugin:@typescript-eslint/recommended", "plugin:react-hooks/recommended"],
  parser: "@typescript-eslint/parser",
  parserOptions: {
    ecmaVersion: "latest",
    sourceType: "module"
  },
  plugins: ["@typescript-eslint"],
  rules: {
    "@typescript-eslint/explicit-module-boundary-types": "off"
  },
  ignorePatterns: ["dist", "node_modules"]
};
