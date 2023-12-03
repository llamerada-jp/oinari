module.exports = {
  mode: "development",

  entry: {
    oinari: "./src/main.ts",
    container: "./src/container/main.ts",
    controller: "./src/controller/main.ts",
    test: "./src/test.ts"
  },
  output: {
    path: `${__dirname}/dist`,
    filename: "[name].js",
    library: "Oinari"
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