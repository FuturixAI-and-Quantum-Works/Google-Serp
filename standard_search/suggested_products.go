package standard_search

import "github.com/PuerkitoBio/goquery"

type SuggestedProduct struct {
	Title  string `json:"title,omitempty"`
	Price  string `json:"price,omitempty"`
	Image  string `json:"image,omitempty"`
	Rating string `json:"rating,omitempty"`
	Seller string `json:"seller,omitempty"`
}

func ExtractSuggestedProducts(doc *goquery.Document) []SuggestedProduct {
	var suggestedProducts []SuggestedProduct

	doc.Find("g-inner-card").Each(func(i int, s *goquery.Selection) {
		product := SuggestedProduct{}

		// Extract product title
		title := s.Find(".gkQHve").Text()
		product.Title = title

		// Extract product price
		price := s.Find("span.lmQWe").Text()
		product.Price = price

		// Extract product image
		image := s.Find(".BYbUcd img").AttrOr("src", "")
		product.Image = image

		// Extract product rating
		rating := s.Find(".z3HNkc").AttrOr("aria-label", "")
		product.Rating = rating

		// Extract product seller
		seller := s.Find(".WJMUdc").Text()
		product.Seller = seller

		// check if product is {} or not
		if product != (SuggestedProduct{}) {

			suggestedProducts = append(suggestedProducts, product)
		}
	})

	return suggestedProducts
}
