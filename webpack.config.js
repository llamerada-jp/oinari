module.exports = {
    mode: "development",

    entry: {
        index: "./src/index.ts",
        worker: "./src/worker.ts"
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