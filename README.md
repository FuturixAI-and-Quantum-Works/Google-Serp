# Google Scraper API

A high-performance web scraping and search API service built with Go, providing endpoints for various search engines, financial data, and general web scraping capabilities.

## Features

- **Search Engine Integration**
  - Google Search
  - Bing Search
  - Google Image Search
  - Google Shopping Search
  
- **Financial Data**
  - Stock price retrieval
  - Live price predictions
  - Stock charts and forecasts
  - Shareholdings information
  
- **Web Scraping**
  - URL scraping with browser automation
  - HTML cleaning and extraction
  - User-agent rotation for reliability
  
- **Additional Features**
  - Redis caching support
  - CORS-enabled REST API
  - Browser automation with Chromedp
  - Python-based scraping utilities

## Prerequisites

- Go 1.23.4 or higher
- Redis (optional, for caching)
- Python 3.x (for additional scraping utilities)
- Chrome/Chromium (for browser automation)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd googlescrapper
```

2. Install Go dependencies:
```bash
go mod download
```

3. Install Python dependencies (if using Python utilities):
```bash
pip install botasaurus
```

## Configuration

The application uses environment variables for configuration:

- `REDIS_ADDR`: Redis server address (default: `localhost:6379`)
- `PORT`: Server port (default: `8000`)

## Running the Server

```bash
go run main.go
```

The server will start on port 8000 by default.

## API Endpoints

### Search Endpoints

- **Standard Search**
  ```
  GET /search/{query}/{location}/{maxResults}/{latitude}/{longitude}/{useCoords}
  ```

- **Bing Search**
  ```
  GET /bing/{query}
  ```

- **Image Search**
  ```
  GET /image/{query}
  ```

- **Shopping Search**
  ```
  GET /shopping/{query}
  ```

### Financial Endpoints

- **Finance Search**
  ```
  GET /finance/{symbol}
  ```

- **Stock Charts**
  ```
  GET /stock/charts
  ```

- **Live Stock Price**
  ```
  GET /stock/live/{tickerId}
  GET /stock/live-price/{tickerId}
  ```

- **Stock Forecast**
  ```
  GET /stock/forecast/{tickerId}
  ```

- **Shareholdings**
  ```
  GET /stock/shareholdings/{tickerId}/{type}
  ```

- **Scrape Stock Data**
  ```
  GET /scrape/{stockIdentifier}
  ```

### Scraping Endpoints

- **Scrape URL**
  ```
  POST /scrape-url
  ```

- **Clean HTML**
  ```
  POST /clean-html
  ```

- **Get HTML from URL**
  ```
  GET /html
  ```

## Usage Examples

### Search Example
```bash
curl "http://localhost:8000/search/golang/us/10/0/0/false"
```

### Stock Price Example
```bash
curl "http://localhost:8000/stock/live-price/AAPL"
```

### Scrape URL Example
```bash
curl -X POST "http://localhost:8000/scrape-url" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

## Python Utilities

The project includes Python scripts for additional scraping capabilities:

- `scrape.py`: Browser-based scraping using Botasaurus
- `get_clean_html.py`: HTML cleaning utilities
- `test_scraper.py`: Test suite for scraping functionality

### Using the Python Scraper
```bash
python scrape.py -u "https://example.com" -f "output.html"
```

## Project Structure

```
.
├── main.go              # Main application entry point
├── go.mod               # Go module definition
├── search/              # Search engine implementations
├── stock/               # Stock and financial data handlers
├── scraper/             # Web scraping utilities
├── browser/             # Browser automation
├── config/              # Configuration utilities
├── cache/               # Caching implementations
├── utils/               # Utility functions
└── output/              # Output directory for scraped data
```

## Development

### Building the Binary
```bash
go build -o googlescrapper
```

### Running Tests
```bash
go test ./...
python test_scraper.py
```

## CORS Configuration

The API is configured with CORS support allowing:
- All origins (`*`)
- Methods: GET, POST, PUT, DELETE, OPTIONS
- Headers: Content-Type, Authorization

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Gorilla Mux](https://github.com/gorilla/mux) for routing
- Uses [Chromedp](https://github.com/chromedp/chromedp) for browser automation
- [Botasaurus](https://github.com/omkarcloud/botasaurus) for Python-based scraping