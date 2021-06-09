const IBGE_PROJECOES_URL =
  "https://servicodados.ibge.gov.br/api/v1/projecoes/populacao/";
const ELASTICSEARCH_URL = "https://imunizacao-es.saude.gov.br/_search";
const ELASTICSEARCH_USERNAME = "imunizacao_public";
const ELASTICSEARCH_PASSWORD = "qlto5t&7r_@+#Tlstigi";
const ELASTICSEARCH_CONTENT = "application/json";
const ELASTICSEARCH_AUTH = btoa(
  `${ELASTICSEARCH_USERNAME}:${ELASTICSEARCH_PASSWORD}`
);
const ELASTICSEARCH_HEADERS = new Headers({
  "Content-Type": ELASTICSEARCH_CONTENT,
  Authorization: `Basic ${ELASTICSEARCH_AUTH}`,
});
const BASE_QUERY: ElasticQuery = (() => {
  const DECODER = new TextDecoder();
  const DATA = Deno.readFileSync("../elastic_query.json");
  return JSON.parse(DECODER.decode(DATA));
})();

const EstadosEnum = {
  RO: 11,
  AC: 12,
  AM: 13,
  RR: 14,
  PB: 25,
  PA: 15,
  AP: 16,
  TO: 17,
  MA: 21,
  PI: 22,
  CE: 23,
  RN: 24,
  PE: 26,
  AL: 27,
  SE: 28,
  BA: 29,
  MG: 31,
  ES: 32,
  RJ: 33,
  SP: 35,
  PR: 41,
  SC: 42,
  RS: 43,
  MS: 50,
  MT: 51,
  GO: 52,
  DF: 53,
} as const;

type UF = keyof typeof EstadosEnum;

//#region Elasticsearch query
interface ElasticQuery {
  size: number;
  aggs: Aggs;
  query?: {
    bool: {
      must: {
        match: {
          estabelecimento_municipio_codigo?: number | string;
          estabelecimento_uf?: UF;
        };
      };
    };
  };
}

interface Aggs {
  filtros: Filtros;
}

interface Filtros {
  filters: FiltrosFilters;
  aggs: FiltrosAggs;
}

interface FiltrosAggs {
  unique_docs: UniqueDocs;
}

interface UniqueDocs {
  cardinality: Cardinality;
}

interface Cardinality {
  field: string;
}

interface FiltrosFilters {
  filters: FiltersFilters;
}

interface FiltersFilters {
  primeira_dose: ADose;
  segunda_dose: ADose;
}

interface ADose {
  match: Match;
}

interface Match {
  vacina_descricao_dose: VacinaDescricaoDose;
}

interface VacinaDescricaoDose {
  query: string;
  operator: string;
}
//#endregion Elasticsearch query

//#region Elasticsearch response
interface ElasticResponse {
  took: number;
  timed_out: boolean;
  _shards: Shards;
  hits: Hits;
  aggregations: Aggregations;
}

interface Shards {
  total: number;
  successful: number;
  skipped: number;
  failed: number;
}

interface Aggregations {
  filtros: Filtros;
}

interface Filtros {
  buckets: Buckets;
}

interface Buckets {
  primeira_dose: ADose;
  segunda_dose: ADose;
}

interface ADose {
  doc_count: number;
  unique_docs: UniqueDocs;
}

interface UniqueDocs {
  value: number;
}

interface Hits {
  total: Total;
  max_score: null;
  hits: any[];
}

interface Total {
  value: number;
  relation: string;
}
//#endregion Elasticsearch response

interface Projecao {
  localidade: string;
  // Horário da projeção no formato dd/MM/yyyy HH:mm:ss
  horario: string;
  projecao: {
    // Projeção populacional
    populacao: number;
    periodoMedio: {
      incrementoPopulacional: string;
      nascimento: string;
      obito: string;
    };
  };
}

interface VacinacaoInfo {
  populacao: number;
  primeira_dose: {
    porcentagem: number;
    total: number;
  };
  segunda_dose: {
    porcentagem: number;
    total: number;
  };
}

type MunicipiosBaseData = {
  [K in UF]: Municipio[];
};

interface Municipio {
  populacao: number;
  codigo_ibge: number;
}

function buildQuery(
  {
    municipioCodigo,
    estadoUF,
  }: {
    municipioCodigo?: number | null;
    estadoUF?: UF | null;
  } = { municipioCodigo: null, estadoUF: null }
): ElasticQuery {
  if (municipioCodigo || estadoUF) {
    const q: ElasticQuery = {
      ...BASE_QUERY,
      query: {
        bool: {
          must: {
            match: {},
          },
        },
      },
    };
    if (municipioCodigo) {
      q.query!.bool.must.match.estabelecimento_municipio_codigo = municipioCodigo
        .toString()
        .slice(0, -1);
    } else if (estadoUF) {
      q.query!.bool.must.match.estabelecimento_uf = estadoUF;
    }
    return q;
  } else return BASE_QUERY;
}

async function unwrapToJSON<T>(request: Promise<Response>) {
  return (await request).json();
}

async function makeElasticRequest(query: ElasticQuery): Promise<Response> {
  return await fetch(ELASTICSEARCH_URL, {
    headers: ELASTICSEARCH_HEADERS,
    method: "POST",
    body: JSON.stringify(query),
  });
}

function treatElasticResponse(
  response: ElasticResponse,
  populacao: number
): VacinacaoInfo {
  const PRIMEIRA_DOSE_COUNT =
    response.aggregations.filtros.buckets.primeira_dose.unique_docs.value;
  const SEGUNDA_DOSE_COUNT =
    response.aggregations.filtros.buckets.segunda_dose.unique_docs.value;

  return {
    populacao,
    primeira_dose: {
      porcentagem: Number.parseFloat(
        ((PRIMEIRA_DOSE_COUNT / populacao) * 100).toFixed(2)
      ),
      total: PRIMEIRA_DOSE_COUNT,
    },
    segunda_dose: {
      porcentagem: Number.parseFloat(
        ((SEGUNDA_DOSE_COUNT / populacao) * 100).toFixed(2)
      ),
      total: SEGUNDA_DOSE_COUNT,
    },
  };
}

async function stageBrasil(): Promise<VacinacaoInfo> {
  const PROJECAO: Projecao = await (await fetch(IBGE_PROJECOES_URL + 0)).json();

  const QUERY = buildQuery();

  const ELASTIC_RESPONSE: ElasticResponse = await unwrapToJSON(
    makeElasticRequest(QUERY)
  );

  return treatElasticResponse(ELASTIC_RESPONSE, PROJECAO.projecao.populacao);
}

async function stageEstados(): Promise<Map<UF, VacinacaoInfo>> {
  const RESULT: Map<UF, VacinacaoInfo> = new Map();
  const ESTADOS_MAP: WeakMap<Promise<Response>, UF> = new WeakMap();

  let requests: Promise<Response>[] = [];

  for (const ID of Object.values(EstadosEnum)) {
    requests.push(fetch(IBGE_PROJECOES_URL + ID));
  }

  const PROJECOES: Projecao[] = await Promise.all(requests.map(unwrapToJSON));

  requests = [];

  for (const UF in EstadosEnum) {
    const QUERY = buildQuery({ estadoUF: UF as UF });
    const REQUEST = makeElasticRequest(QUERY);
    ESTADOS_MAP.set(REQUEST, UF as UF);
    requests.push(REQUEST);
  }

  const FACTORY: Promise<void>[] = [];

  for (let index = 0, length = requests.length; index < length; index++) {
    FACTORY.push(
      (async (index: number) => {
        const UF = ESTADOS_MAP.get(requests[index]);
        if (!UF) return;
        const ELASTIC_RESPONSE = await unwrapToJSON(requests[index]);
        RESULT.set(
          UF,
          treatElasticResponse(
            ELASTIC_RESPONSE,
            PROJECOES[index].projecao.populacao
          )
        );
        return;
      })(index)
    );
  }

  await Promise.all(FACTORY);

  return RESULT;
}

async function stageMunicipios(): Promise<Map<UF, VacinacaoInfo[]>> {
  const RESULT: Map<UF, VacinacaoInfo[]> = new Map();
  for (const UF of Object.keys(EstadosEnum)) {
    RESULT.set(UF as UF, []);
  }

  const ESTADOS_MAP: WeakMap<Promise<Response>, UF> = new WeakMap();
  const MUNICIPIO_MAP: WeakMap<Promise<Response>, Municipio> = new WeakMap();

  const MUNICIPIOS_BASE: MunicipiosBaseData = await (async () => {
    const DECODER = new TextDecoder();
    const DATA = await Deno.readFile("../municipios.json");
    return JSON.parse(DECODER.decode(DATA));
  })();

  const ELASTIC_REQUESTS: Promise<Response>[] = [];

  for (const [ESTADO_UF, MUNICIPIOS] of Object.entries(MUNICIPIOS_BASE)) {
    for (const MUNICIPIO of MUNICIPIOS) {
      const QUERY = buildQuery({ municipioCodigo: MUNICIPIO.codigo_ibge });
      const REQUEST = makeElasticRequest(QUERY);
      MUNICIPIO_MAP.set(REQUEST, MUNICIPIO);
      ESTADOS_MAP.set(REQUEST, ESTADO_UF as UF);
      ELASTIC_REQUESTS.push(REQUEST);
    }
  }

  const FACTORY: Promise<void>[] = [];

  for (const REQUEST of ELASTIC_REQUESTS) {
    FACTORY.push(
      (async (REQUEST: Promise<Response>) => {
        const UF = ESTADOS_MAP.get(REQUEST);
        const MUNICIPIO = MUNICIPIO_MAP.get(REQUEST);
        if (UF && MUNICIPIO) {
          const RESPONSE: ElasticResponse = await unwrapToJSON(REQUEST);
          RESULT.get(UF)?.push(
            treatElasticResponse(RESPONSE, MUNICIPIO.populacao)
          );
        }
        return;
      })(REQUEST)
    );
  }

  return RESULT;
}

export {};
