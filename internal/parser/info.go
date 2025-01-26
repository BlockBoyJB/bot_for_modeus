package parser

import "net/http"

const (
	findBuildings = "/info/buildings"
)

type Building struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	SearchUrl string `json:"search_url"` // https ссылка на яндекс карты
}

func (p *parser) FindBuildings() ([]Building, error) {
	resp, err := p.makeRequest(http.MethodGet, findBuildings, nil)
	if err != nil {
		return nil, err
	}
	var result []Building
	if err = parseBody(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}
