import argparse
from botasaurus.browser import browser, Driver


@browser(reuse_driver=True, cache=True)
def scrape_data(driver: Driver, link):
    driver.get(link["url"])
    # save to file
    with open(link["filename"], "w") as f:
        f.write(driver.page_html)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Scrape data from a URL")
    parser.add_argument("-u", "--url", help="URL to scrape", required=True)
    parser.add_argument(
        "-f", "--filename", help="Filename to save the HTML", required=True
    )
    args = parser.parse_args()

    link = {"url": args.url, "filename": args.filename}

    scrape_data([link])
