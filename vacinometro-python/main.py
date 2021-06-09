import asyncio
import json
from .utils import (
    asnyc_aiohttp_get_all,
    asnyc_aiohttp_get_all_elastic,
    fetch_single,
    fetch_vacinas_total,
)

ESTADOS_UF = {
    11: "RO",
    12: "AC",
    13: "AM",
    14: "RR",
    25: "PB",
    15: "PA",
    16: "AP",
    17: "TO",
    21: "MA",
    22: "PI",
    23: "CE",
    24: "RN",
    26: "PE",
    27: "AL",
    28: "SE",
    29: "BA",
    31: "MG",
    32: "ES",
    33: "RJ",
    35: "SP",
    41: "PR",
    42: "SC",
    43: "RS",
    50: "MS",
    51: "MT",
    52: "GO",
    53: "DF",
}

IBGE_PROJECOES_URL = "https://servicodados.ibge.gov.br/api/v1/projecoes/populacao/"

def treat_elastic_response(elastic_response, populacao: int):
    primeira_dose_count: int = elastic_response["aggregations"]["filtros"]["buckets"][
        "primeira_dose"
    ]["unique_docs"]["value"]

    segunda_dose_count: int = elastic_response["aggregations"]["filtros"]["buckets"][
        "segunda_dose"
    ]["unique_docs"]["value"]

    return {
        "populacao": populacao,
        "primeira_dose": {
            "total": primeira_dose_count,
            "porcentagem": round(primeira_dose_count / populacao * 100, 2),
        },
        "segunda_dose": {
            "total": segunda_dose_count,
            "porcentagem": round(segunda_dose_count / populacao * 100, 2),
        },
    }


def stage_vacinometros():
    loop = asyncio.get_event_loop()
    # Brasil = 0 ou BR
    projecao_response = loop.run_until_complete(fetch_single(IBGE_PROJECOES_URL.join("0")))

    elastic_response, _ = loop.run_until_complete(fetch_vacinas_total())

    populacao: int = projecao_response["projecao"]["populacao"]

    result = treat_elastic_response(elastic_response, populacao)

    return result


def stage_estados():
    result = {}

    # Estados 11 - 53
    urls = [IBGE_PROJECOES_URL.join(str(codigo)) for codigo in ESTADOS_UF.keys()]

    # Pega projeções de população
    projecoes_responses = asnyc_aiohttp_get_all(urls)

    projecoes_populacoes: dict[str, int] = {}

    for projecao in projecoes_responses:
        ibge_id = int(projecao["localidade"])
        populacao: int = projecao["projecao"]["populacao"]

        uf = ESTADOS_UF.get(ibge_id)

        projecoes_populacoes[uf] = populacao

    gathered_elastic_responses = asnyc_aiohttp_get_all_elastic(
        estados_uf=list(ESTADOS_UF.values())
    )

    # TODO: Implement async append
    for elastic_response, uf in gathered_elastic_responses:
        populacao = projecoes_populacoes[uf]
        result[uf] = treat_elastic_response(elastic_response, populacao)

    return result

def stage_municipios():
    municipios_json = json.load(open("../municipios.json", "r"))

    gathered_municipios = {}
    result = {}

    for uf in municipios_json:
        result[uf] = []
        codigos = [dado["codigo_ibge"] for dado in municipios_json[uf]]
        gathered_municipios[uf] = asnyc_aiohttp_get_all_elastic(ibge_codigos=codigos)

    # TODO: Implement async append
    for uf, gathered_responses in gathered_municipios.items():
        for elastic_response, codigo in gathered_responses:
            item = next(row for row in municipios_json[uf] if row["codigo_ibge"] == codigo)
            populacao = int(item["populacao"])
            result[uf].append(treat_elastic_response(elastic_response, populacao))

    return result


if __name__ == "__main__":
    stage_vacinometros()

    stage_estados()

    stage_municipios()

# Or this, if needed:
# async def main():
#     stage_vacinometros()

#     stage_estados()

#     stage_municipios()

# asyncio.run(main())