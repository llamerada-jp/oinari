module.exports = {
  mode: "development",

  entry: {
    index: "./src/index.ts",
    container: "./src/container/index.ts",
    controller: "./src/controller/index.ts",
    test: "./src/test.ts"
  },
  output: {
    path: `${__dirname}/dist`,
    filename: "[name].js"
  },
  module: {
    rules: [{
      test: /\.ts$/,
      use: "ts-loader"
    }]
  },
  resolve: {
    extensions: [".ts", ".js"]
  }
};