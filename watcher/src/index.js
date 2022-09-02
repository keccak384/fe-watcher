const { By, Key, Builder } = require("selenium-webdriver");
const chrome = require("selenium-webdriver/chrome");
require("chromedriver");

// session in unix
function testFunc(session) {
    const checkContents = () => {
        const safeScripts = [
            "(function(){window['__CF$cv$params']",
            "(function(){var js = \"window['__CF$cv$params']=",
            '"dark"===function()',
            "!function(e){function t(t){for(var a,o,n=t[0],i=t[1],p=t",
        ];
        const elements = document.getElementsByTagName("script");
        const contents = [...elements]
            .map((e) => e.innerHTML)
            .filter(
                (s) => s.length > 0 && safeScripts.every((ss) => !s.startsWith(ss))
            );

        return contents;
    };

    const checkSources = () => {
        const safeSources = [
            "/cdn-cgi/bm/cv/669835187/api.js",
            "/datafeeds/udf/dist/bundle.js",
            "/static/js/26.787fb90a.chunk.js",
            "/static/js/24.e6826c03.chunk.js",
            "/static/js/32.8834e2d3.chunk.js",
            "/charting_library/charting_library.standalone.js",
            "/static/js/ethers.7bd0932d.chunk.js",
            "/static/js/main.56a19174.chunk.js",
            "https://static.cloudflareinsights.com/beacon.min.js/v652eace1692a40cfa3763df669d7439c1639079717194",
        ];

        const elements = document.getElementsByTagName("script");
        const srcs = [...elements]
            .map((e) => e.getAttribute("src"))
            .filter(Boolean)
            .filter((s) => !safeSources.includes(s));

        return srcs;
    };

    setInterval(async () => {
        const contents = checkContents();
        const sources = checkSources();

        if (contents.length > 0 || sources.length > 0) {
            if (!window.wth) {
                window.eth = [];
            }

            window.eth.push({
                contents,
                sources,
            });

            const req = await fetch(`${process.env.API_URL}/api/logs`, {
                method: "POST",
                body: JSON.stringify({
                    contents: contents,
                    sources: sources,
                    session: session,
                }),
            });
            console.log({ req });
        }
    }, 50);
}

const scriptStr = `;(() => {
    const session = Math.floor(Date.now() / 1000);
    ${testFunc.toString()}
    testFunc(session);
})()
`;

// console.log({ scriptStr });

const sleep = (delay) => {
    return new Promise((resolve) => {
        setTimeout(resolve, delay);
    });
};

const screen = {
    width: 1280,
    height: 720,
};

const main = async () => {
    // To wait for browser to build and launch properly
    let driver = await new Builder()
        .forBrowser("chrome")
        .setChromeOptions(new chrome.Options().headless().windowSize(screen))
        .build();

    await driver.get("http://kyberswap.com");
    driver.executeScript(scriptStr);

    setInterval(async () => {
        await driver.navigate().refresh();
        driver.executeScript(scriptStr);
        await driver.sleep(5_000);
        await driver.findElement(By.className("SearchInput-sc-1bq2i12-1")).click();
    }, 10_000);

    // It is always a safe practice to quit the browser after execution
    // await driver.quit();
};

main();
