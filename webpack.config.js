module.exports = {
    mode: "development",
  
    entry: "./src/index.ts",
    output: {
      path: `${__dirname}/dist`,
      filename: "oinari.js"
    },
    module: {
      rules: [
        {
          test: /\.ts$/,
          use: "ts-loader"
        }
      ]
    },
    resolve: {
      extensions: [".ts", ".js"]
    }
  };