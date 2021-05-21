package main

type result struct {
	item []*ContentItem
	err  error
}
type ResultChannels []chan result

// Parameters is a typed structure to make it
// harder to confuse the offset and count parameters to the functions
type Parameters struct {
	ip     string
	offset int
	count  int
}

// GetContentItems does concurrent requests for each config respecting there fallback if existing,
// it returns ordered slice, any error will result in subsequent values being discarded
func GetContentItems(a App, p Parameters) []ContentItem {
	return getResults(a, p).toSlice()
}

// getResult runs the request in separate goroutines,
// it returns an ordered slice of channels
func getResults(a App, p Parameters) ResultChannels {
	length := len(a.Config)
	if length == 0 {
		return nil
	}

	resultChannels := make(ResultChannels, 0, p.count)
	for i := 0; i < p.count; i++ {
		resultChan := make(chan result)
		resultChannels = append(resultChannels, resultChan)
		go func(index int) {
			resultChan <- getRequest(p.ip, getClients(a, index%length))
			close(resultChan)
		}(i + p.offset)
	}
	return resultChannels
}

// Makes sure channels are read in order
// If any channels return error it breaks and returns the proceeding ContentItems
func (r ResultChannels) toSlice() []ContentItem {
	var contentItems = make([]ContentItem, 0, len(r))
	for _, c := range r {
		result := <-c
		if result.err != nil || result.item[0] == nil {
			break
		}
		contentItems = append(contentItems, *result.item[0])
	}
	return contentItems
}

// getClients returns the clients from the config
// this will be the .Type and if present the .Fallback
func getClients(a App, index int) []Client {
	c := a.Config[index]
	clients := append(make([]Client, 0, 2), a.ContentClients[c.Type])
	if c.Fallback != nil {
		clients = append(clients, a.ContentClients[*c.Fallback])
	}
	return clients
}

// getRequest makes the request to GetContent testing each client until success
// and then returns a result of the []*ContentItem and possible error
func getRequest(ip string, clients []Client) result {
	var item []*ContentItem
	var err error
	for _, c := range clients {
		item, err = c.GetContent(ip, 1)
		if err == nil {
			break
		}
	}
	return result{item, err}
}
