#!/usr/bin/env python3
"""
Test script for the URL scraper API
Sends a URL to the API and displays the Markdown response in a rendered format
"""

import argparse
import json
import requests
import sys
import tempfile
import os
import subprocess
import webbrowser

# Default API endpoint
DEFAULT_ENDPOINT = "http://localhost:8000/scrape-url"

def send_url_to_scraper(url, endpoint=DEFAULT_ENDPOINT):
    """Send a URL to the scraper API and get the Markdown content back"""
    payload = {"url": url}
    headers = {"Content-Type": "application/json"}
    
    try:
        response = requests.post(endpoint, json=payload, headers=headers)
        response.raise_for_status()  # Raise exception for non-200 responses
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error making request: {e}")
        sys.exit(1)

def save_as_markdown(content, output_file=None):
    """Save content to a Markdown file"""
    if output_file is None:
        # Create a temporary file with .md extension
        fd, output_file = tempfile.mkstemp(suffix='.md')
        os.close(fd)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(content)
    
    return output_file

def render_markdown(markdown_file):
    """Render Markdown in a way appropriate for the platform"""
    # First try to open with a Markdown viewer if available
    try:
        # Check if pandoc is installed
        if subprocess.run(['which', 'pandoc'], stdout=subprocess.PIPE, stderr=subprocess.PIPE).returncode == 0:
            # Convert to HTML and open in browser
            html_file = markdown_file.replace('.md', '.html')
            subprocess.run(['pandoc', markdown_file, '-o', html_file])
            webbrowser.open(f'file://{os.path.abspath(html_file)}')
            return
    except Exception:
        pass
    
    # Fallback: Try to use VS Code if available
    try:
        subprocess.run(['code', markdown_file])
        return
    except Exception:
        pass
    
    # Last resort: just print to terminal and provide the file path
    with open(markdown_file, 'r', encoding='utf-8') as f:
        print(f.read())
    print(f"\nMarkdown saved to: {markdown_file}")

def main():
    parser = argparse.ArgumentParser(description='Test the URL scraper API')
    parser.add_argument('url', help='URL to scrape')
    parser.add_argument('--endpoint', default=DEFAULT_ENDPOINT, 
                        help=f'API endpoint (default: {DEFAULT_ENDPOINT})')
    parser.add_argument('--output', '-o', help='Output file (default: temporary file)')
    parser.add_argument('--raw', '-r', action='store_true', 
                        help='Show raw JSON response instead of rendering Markdown')
    
    args = parser.parse_args()
    
    # Send URL to the scraper
    print(f"Scraping URL: {args.url}")
    response_data = send_url_to_scraper(args.url, args.endpoint)
    
    if args.raw:
        # Print the raw response
        print(json.dumps(response_data, indent=2))
        return
    
    # Extract Markdown content
    if 'markdown' in response_data:
        markdown_content = response_data['markdown']
        
        # Save to file
        output_file = save_as_markdown(markdown_content, args.output)
        
        # Render the Markdown
        print(f"Successfully scraped content from {args.url}\n")
        render_markdown(output_file)
    else:
        print("Error: Response does not contain Markdown content")
        print(json.dumps(response_data, indent=2))

if __name__ == "__main__":
    main()