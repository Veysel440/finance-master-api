package http

type API struct {
	Auth   *AuthHandlers
	H      *Handlers
	CatH   *CatalogHandlers
	Rates  *RatesHandlers
	Secret []byte
}
