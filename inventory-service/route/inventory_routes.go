package route

import (
	"net/http"

	"example.com/go-commerce-platform/inventory-service/handler"
)

func RegisterInventoryRoutes(h *handler.InventoryHandler) {
	http.HandleFunc("/products", h.Products)
	http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > len("/products/") &&
			r.URL.Path[len("/products/"):] != "" &&
			len(r.URL.Path) >= len("/products/stock") &&
			r.URL.Path[len(r.URL.Path)-6:] == "/stock" {
			h.Stock(w, r)
			return
		}

		h.ProductByID(w, r)
	})
}
