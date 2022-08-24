package gout

// GET send HTTP GET method
func GET(url string) *DataFlow {
	return New().GET(url)
}

// POST send HTTP POST method
func POST(url string) *DataFlow {
	return New().POST(url)
}

// PUT send HTTP PUT method
func PUT(url string) *DataFlow {
	return New().PUT(url)
}

// DELETE send HTTP DELETE method
func DELETE(url string) *DataFlow {
	return New().DELETE(url)
}
