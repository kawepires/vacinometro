package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"unicode/utf8"
)

// #region Constants and variables
const (
	IBGE_PROJECOES_URL     string = "https://servicodados.ibge.gov.br/api/v1/projecoes/populacao/%d"
	ELASTICSEARCH_URL      string = "https://imunizacao-es.saude.gov.br/_search"
	ELASTICSEARCH_USERNAME string = "imunizacao_public"
	ELASTICSEARCH_PASSWORD string = "qlto5t&7r_@+#Tlstigi"
	ELASTICSEARCH_CONTENT  string = "application/json"
)

var elasticSearchAuth = base64.RawStdEncoding.EncodeToString([]byte(ELASTICSEARCH_USERNAME + ":" + ELASTICSEARCH_PASSWORD))

var elasticsearchBaseQuery BaseRequestPayload
var estadosBaseData EstadosBaseData

// Definição das UFs dos estados e seus respectivos códigos do IBGE
var estadosMapa = map[string]uint8{
	"RO": 11,
	"AC": 12,
	"AM": 13,
	"RR": 14,
	"PB": 25,
	"PA": 15,
	"AP": 16,
	"TO": 17,
	"MA": 21,
	"PI": 22,
	"CE": 23,
	"RN": 24,
	"PE": 26,
	"AL": 27,
	"SE": 28,
	"BA": 29,
	"MG": 31,
	"ES": 32,
	"RJ": 33,
	"SP": 35,
	"PR": 41,
	"SC": 42,
	"RS": 43,
	"MS": 50,
	"MT": 51,
	"GO": 52,
	"DF": 53,
}

// #endregion Constants and variables

// #region Types

// #region Elastic request payload structs

// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

type BaseRequestPayload struct {
	Size int        `json:"size"`
	Aggs AggsOutter `json:"aggs"`
}

type VacinaDescricaoDose struct {
	Query    string `json:"query"`
	Operator string `json:"operator"`
}

type Match struct {
	VacinaDescricaoDose VacinaDescricaoDose `json:"vacina_descricao_dose"`
}

type PrimeiraDose struct {
	Match Match `json:"match"`
}

type SegundaDose struct {
	Match Match `json:"match"`
}

type FiltersInner struct {
	PrimeiraDose PrimeiraDose `json:"primeira_dose"`
	SegundaDose  SegundaDose  `json:"segunda_dose"`
}

type FiltersOutter struct {
	Filters FiltersInner `json:"filters"`
}

type Cardinality struct {
	Field string `json:"field"`
}

type UniqueDocs struct {
	Cardinality Cardinality `json:"cardinality"`
}

type AggsInner struct {
	UniqueDocs UniqueDocs `json:"unique_docs"`
}

type Filtros struct {
	Filters FiltersOutter `json:"filters"`
	Aggs    AggsInner     `json:"aggs"`
}

type AggsOutter struct {
	Filtros Filtros `json:"filtros"`
}

// #region Specialized requests
type MunicipioRequestPayload struct {
	BaseRequestPayload
	MunicipioRequestQueryPayload
}

type EstadoRequestPayload struct {
	BaseRequestPayload
	EstadoRequestQueryPayload
}

type MunicipioRequestQueryPayload struct {
	Query QueryEstabelecimento `json:"query"`
}

type EstadoRequestQueryPayload struct {
	Query QueryEstado `json:"query"`
}

type QueryEstabelecimento struct {
	Bool BoolEstabelecimento `json:"bool"`
}

type BoolEstabelecimento struct {
	Must MustMatchEstabelecimento `json:"must"`
}

type MustMatchEstabelecimento struct {
	Match MatchEstabelecimento `json:"match"`
}

type MatchEstabelecimento struct {
	EstabelecimentoMunicipioCodigo uint32 `json:"estabelecimento_municipio_codigo"`
}

type QueryEstado struct {
	Bool BoolEstado `json:"bool"`
}

type BoolEstado struct {
	Must MustMatchEstado `json:"must"`
}

type MustMatchEstado struct {
	Match MatchEstado `json:"match"`
}

type MatchEstado struct {
	EstadoEstabelecimentoUf string `json:"estabelecimento_uf"`
}

// #endregion Specialized requests

// #region Responses
type ElasticResponse struct {
	Aggregations Aggregations `json:"aggregations"`
}
type Aggregations struct {
	Filtros ReponseFiltros `json:"filtros"`
}
type ReponseFiltros struct {
	Buckets Buckets `json:"buckets"`
}
type Buckets struct {
	PrimeiraDose ResponseDose `json:"primeira_dose"`
	SegundaDose  ResponseDose `json:"segunda_dose"`
}

type ResponseDose struct {
	DocCount   uint32             `json:"doc_count"`
	UniqueDocs ResponseUniqueDocs `json:"unique_docs"`
}

type ResponseUniqueDocs struct {
	Value uint32 `json:"value"`
}

// #endregion Responses

// #endregion Elastic request payload structs

type Municipio struct {
	CodigoIbge uint32 `json:"codigo_ibge"`
	Populacao  uint32 `json:"populacao"`
}

type EstadosBaseData map[string][]Municipio

type Projecao struct {
	Populacao uint32 `json:"populacao"`
}

type ProjecaoApi struct {
	Projecao Projecao `json:"projecao"`
}

type RegiaoVacinaInfos struct {
	Populacao    uint32    `json:"populacao"`
	PrimeiraDose DoseInfos `json:"primeira_dose"`
	SegundaDose  DoseInfos `json:"segunda_dose"`
}

type DoseInfos struct {
	Total       uint32  `json:"total"`
	Porcentagem float32 `json:"porcentagem"`
}

type BuildQueryParams struct {
	EstadoUF        string
	MunicipioCodigo uint32
}

// #endregion Types

// #region Functions

func handleReadError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func handleUnmarshalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getPercentage(populacao uint32, primeiraDoseCount uint32, segundaDoseCount uint32) (vacinadosPrimeiraDose float32, vacinadosSegundaDose float32) {
	vacinadosPrimeiraDose = float32(primeiraDoseCount) / float32(populacao) * 100
	vacinadosSegundaDose = float32(segundaDoseCount) / float32(populacao) * 100

	return vacinadosPrimeiraDose, vacinadosSegundaDose
}

func removeLastChar(str string) string {
	for len(str) > 0 {
		_, size := utf8.DecodeLastRuneInString(str)
		return str[:len(str)-size]
	}
	return str
}

func buildQuery(params BuildQueryParams) (interface{}, error) {

	if params.MunicipioCodigo != 0 {
		stringId := removeLastChar(fmt.Sprint(params.MunicipioCodigo))
		ibgeCodUint32, err := strconv.ParseUint(stringId, 10, 32)

		if err != nil {
			return nil, err
		}

		q := MunicipioRequestPayload{
			BaseRequestPayload: elasticsearchBaseQuery,
			MunicipioRequestQueryPayload: MunicipioRequestQueryPayload{
				Query: QueryEstabelecimento{
					Bool: BoolEstabelecimento{
						Must: MustMatchEstabelecimento{
							Match: MatchEstabelecimento{
								EstabelecimentoMunicipioCodigo: uint32(ibgeCodUint32),
							},
						},
					},
				},
			},
		}
		return q, nil
	} else if params.EstadoUF != "" {
		q := EstadoRequestPayload{
			BaseRequestPayload: elasticsearchBaseQuery,
			EstadoRequestQueryPayload: EstadoRequestQueryPayload{
				Query: QueryEstado{
					Bool: BoolEstado{
						Must: MustMatchEstado{
							Match: MatchEstado{
								EstadoEstabelecimentoUf: params.EstadoUF,
							},
						},
					},
				},
			},
		}
		return q, nil
	}
	return elasticsearchBaseQuery, nil
}

func makeElasticRequest(payload []byte) (*http.Response, error) {
	body := bytes.NewReader(payload)
	req, err := http.NewRequest("GET", ELASTICSEARCH_URL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", ELASTICSEARCH_CONTENT)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", elasticSearchAuth))

	return http.DefaultClient.Do(req)
}

func getVacinacoes(query interface{}) (*ElasticResponse, error) {
	payloadBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	resp, err := makeElasticRequest(payloadBytes)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var elasticResp *ElasticResponse

	if err := json.NewDecoder(resp.Body).Decode(&elasticResp); err != nil {
		return nil, err
	}

	return elasticResp, nil
}

func treatElasticResponse(elasticResponse *ElasticResponse, populacao uint32) *RegiaoVacinaInfos {
	primeiraDoseCount := elasticResponse.Aggregations.Filtros.Buckets.PrimeiraDose.UniqueDocs.Value
	segundaDoseCount := elasticResponse.Aggregations.Filtros.Buckets.SegundaDose.UniqueDocs.Value

	vacinadosPrimeiraDose, vacinadosSegundaDose := getPercentage(populacao, primeiraDoseCount, segundaDoseCount)

	result := &RegiaoVacinaInfos{
		Populacao: populacao,
		PrimeiraDose: DoseInfos{
			Total:       primeiraDoseCount,
			Porcentagem: vacinadosPrimeiraDose,
		},
		SegundaDose: DoseInfos{
			Total:       segundaDoseCount,
			Porcentagem: vacinadosSegundaDose,
		},
	}

	return result
}

func getProjecao(id uint8) (projecao *ProjecaoApi, err error) {
	url := fmt.Sprintf(IBGE_PROJECOES_URL, id)

	ibgeApiResp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer ibgeApiResp.Body.Close()

	if err := json.NewDecoder(ibgeApiResp.Body).Decode(&projecao); err != nil {
		return nil, err
	} else {
		return projecao, nil
	}
}

func stageBrasil() (*RegiaoVacinaInfos, error) {

	projecaoApi, err := getProjecao(0)

	if err != nil {
		return nil, err
	}

	var populacaoBrasil = projecaoApi.Projecao.Populacao

	payload, err := buildQuery(BuildQueryParams{})

	if err != nil {
		return nil, err
	}

	elasticResponse, err := getVacinacoes(payload)

	if err != nil {
		return nil, err
	}

	infos := treatElasticResponse(elasticResponse, populacaoBrasil)

	return infos, nil
}

func stageEstados() (*map[string]*RegiaoVacinaInfos, error) {

	var wg sync.WaitGroup
	var mutex sync.Mutex

	estadosProjecoesMap := make(map[string]*RegiaoVacinaInfos, len(estadosMapa))

	for uf, codIbge := range estadosMapa {
		wg.Add(1)
		go func(uf string, id uint8, estadosProjecoesMap *map[string]*RegiaoVacinaInfos) {
			projecao, err := getProjecao(id)
			if err != nil {
				log.Fatal(err)
				return
			}
			mutex.Lock()
			(*estadosProjecoesMap)[uf] = &RegiaoVacinaInfos{
				Populacao: projecao.Projecao.Populacao,
			}
			mutex.Unlock()
			wg.Done()
		}(uf, codIbge, &estadosProjecoesMap)
	}

	wg.Wait()

	for uf := range estadosMapa {
		wg.Add(1)
		go func(uf string, estadosProjecoesMap *map[string]*RegiaoVacinaInfos) {
			defer wg.Done()

			payload, err := buildQuery(BuildQueryParams{EstadoUF: uf})

			if err != nil {
				log.Fatal(err)
				return
			}

			elasticResponse, err := getVacinacoes(payload)

			if err != nil {
				return
			}

			populacao := (*estadosProjecoesMap)[uf].Populacao
			infos := treatElasticResponse(elasticResponse, populacao)

			mutex.Lock()

			(*estadosProjecoesMap)[uf] = infos

			mutex.Unlock()
		}(uf, &estadosProjecoesMap)
	}

	wg.Wait()

	return &estadosProjecoesMap, nil
}

func stageMunicipios() (*map[string][]*RegiaoVacinaInfos, error) {

	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Prepare union
	municipiosProjecoesMap := make(map[string][]*RegiaoVacinaInfos, len(estadosMapa))

	for uf, municipios := range estadosBaseData {
		municipiosProjecoesMap[uf] = make([]*RegiaoVacinaInfos, len(municipios))
	}

	for uf, municipios := range estadosBaseData {
		wg.Add(len(municipios))
		for index, municipio := range municipios {
			go func(position int, uf string, municipio *Municipio, municipiosProjecoesMap *map[string][]*RegiaoVacinaInfos) {
				defer wg.Done()

				payload, err := buildQuery(BuildQueryParams{MunicipioCodigo: municipio.CodigoIbge})

				if err != nil {
					return
				}

				elasticResponse, err := getVacinacoes(payload)

				if err != nil {
					return
				}

				populacao := municipio.Populacao
				infos := treatElasticResponse(elasticResponse, populacao)

				mutex.Lock()

				(*municipiosProjecoesMap)[uf][position] = infos

				mutex.Unlock()
			}(index, uf, &municipio, &municipiosProjecoesMap)
		}
		wg.Wait()
	}

	return &municipiosProjecoesMap, nil
}

func init() {
	municipiosData, err := ioutil.ReadFile("../municipios.json")
	handleReadError(err)

	err = json.Unmarshal(municipiosData, &estadosBaseData)
	handleUnmarshalError(err)

	baseQueryData, err := ioutil.ReadFile("../elastic_query.json")
	handleReadError(err)

	err = json.Unmarshal(baseQueryData, &elasticsearchBaseQuery)
	handleUnmarshalError(err)
}

func main() {

	resultBrasil, errBrasil := stageBrasil()

	if errBrasil != nil {
		log.Fatal(errBrasil)
	} else {
		log.Println(resultBrasil)
	}

	resultEstados, errEstados := stageEstados()

	if errEstados != nil {
		log.Fatal(errEstados)
	} else {
		fmt.Println((*resultEstados)["AM"])
	}

	resultMunicipios, errMunicipios := stageMunicipios()

	if errMunicipios != nil {
		log.Fatal(errMunicipios)
	} else {
		fmt.Printf("%v", *(*resultMunicipios)["AM"][0])
	}

}

// #endregion Functions
