#!/usr/bin/env python3
"""
Script to fetch clean HTML from a URL and save it to a file
This is useful for developing new specialized scrapers
"""

import argparse
import json
import requests
import sys
import os
from bs4 import BeautifulSoup

# Default API endpoint
DEFAULT_ENDPOINT = "http://localhost:8000/clean-html"

def get_clean_html(url, endpoint=DEFAULT_ENDPOINT):
    """Fetch clean HTML from the API endpoint"""
    payload = {"url": url}
    headers = {"Content-Type": "application/json"}
    
    try:
        response = requests.post(endpoint, json=payload, headers=headers)
        
        # Handle specialized scraper case
        if response.status_code == 409:
            print(f"Error: {response.text}")
            print("A specialized scraper already exists for this URL.")
            return None
        
        response.raise_for_status()  # Raise exception for other non-200 responses
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error making request: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response: {e.response.text}")
        sys.exit(1)

def pretty_print_html(html_content, output_file=None):
    """Format and save or print HTML content"""
    try:
        # Use BeautifulSoup to pretty-print the HTML
        soup = BeautifulSoup(html_content, 'html.parser')
        pretty_html = soup.prettify()
        
        if output_file:
            with open(output_file, 'w', encoding='utf-8') as f:
                f.write(pretty_html)
            print(f"HTML saved to: {output_file}")
        else:
            print(pretty_html)
    except Exception as e:
        print(f"Error formatting HTML: {e}")
        
        # Fallback to raw HTML
        if output_file:
            with open(output_file, 'w', encoding='utf-8') as f:
                f.write(html_content)
            print(f"Raw HTML saved to: {output_file}")
        else:
            print(html_content)

def extract_domain(url):
    """Extract domain name from URL for use in filenames"""
    url = url.replace('https://', '').replace('http://', '')
    domain = url.split('/')[0]
    return domain.replace('.', '_')

def main():
    parser = argparse.ArgumentParser(description='Fetch clean HTML from a URL')
    parser.add_argument('url', help='URL to fetch clean HTML from')
    parser.add_argument('--endpoint', default=DEFAULT_ENDPOINT, 
                        help=f'API endpoint (default: {DEFAULT_ENDPOINT})')
    parser.add_argument('--output', '-o', help='Output file (default: domain_clean.html)')
    
    args = parser.parse_args()
    
    # Generate default output filename if not specified
    if not args.output:
        domain = extract_domain(args.url)
        args.output = f"{domain}_clean.html"
    
    # Fetch clean HTML
    print(f"Fetching clean HTML from: {args.url}")
    response_data = get_clean_html(args.url, args.endpoint)
    
    if response_data and 'html' in response_data:
        print(f"Successfully fetched HTML from {args.url}")
        pretty_print_html(response_data['html'], args.output)
    else:
        print("Error: Response does not contain HTML content")

if __name__ == "__main__":
    main()