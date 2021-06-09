import asyncio
import json
from typing import List, Optional
from asyncio.events import AbstractEventLoop
from aiohttp.client import ClientSession
import aiohttp

ELASTIC_BASE_URL = "https://imunizacao-es.saude.gov.br/_search"
ELASTIC_AUTH = ("imunizacao_public", "qlto5t&7r_@+#Tlstigi")
ELASTIC_HEADERS = {"content-type": "application/json"}
BASE_QUERY = json.load(open("../elastic_query.json", "r", encoding="utf-8"))


def build_query(
    estado_uf: Optional[str] = None, municipio_cod_ibge: Optional[int] = None
):
    if municipio_cod_ibge or estado_uf:
        base_query = BASE_QUERY.copy()
        if municipio_cod_ibge:
            base_query["query"] = {
                "bool": {
                    "must": {
                        "match": {
                            # O Elasticsearch remove o último digito.
                            "estabelecimento_municipio_codigo": int(
                                str(municipio_cod_ibge)[:-1]
                            )
                        }
                    }
                }
            }
        elif estado_uf:
            base_query["query"] = {
                "bool": {"must": {"match": {"estabelecimento_uf": estado_uf}}}
            }
        return base_query

    return BASE_QUERY


async def fetch_elastic(
    session: ClientSession,
    ibge_codigo: Optional[int] = None,
    estado_uf: Optional[str] = None,
):
    """
    Asynchronous get request
    """
    async with session.request(
        "GET",
        ELASTIC_BASE_URL,
        data=json.dumps(
            build_query(municipio_cod_ibge=ibge_codigo, estado_uf=estado_uf)
        ),
    ) as response:
        response_json = await response.json()
        return (response_json, ibge_codigo or estado_uf or None)


async def fetch_many_elastic(
    loop: AbstractEventLoop,
    ibge_codigos: Optional[List[int]] = None,
    estados_uf: Optional[List[str]] = None,
):
    """
    Many asynchronous get requests, gathered
    """
    async with aiohttp.ClientSession(
        auth=aiohttp.BasicAuth(ELASTIC_AUTH[0], ELASTIC_AUTH[1]),
        headers=ELASTIC_HEADERS,
    ) as session:
        if ibge_codigos is not None:
            tasks = [
                loop.create_task(fetch_elastic(session, ibge_codigo=codigo))
                for codigo in ibge_codigos
            ]
        if estados_uf is not None:
            tasks = [
                loop.create_task(fetch_elastic(session, estado_uf=uf))
                for uf in estados_uf
            ]
        if estados_uf is None and ibge_codigos is None:
            raise Exception("Nenhuma lista providenciada.")
        return await asyncio.gather(*tasks)


def asnyc_aiohttp_get_all_elastic(
    ibge_codigos: Optional[List[int]] = None, estados_uf: Optional[List[str]] = None
):
    """
    Performs asynchronous get requests
    """
    loop = asyncio.get_event_loop()
    return loop.run_until_complete(
        fetch_many_elastic(loop, ibge_codigos=ibge_codigos, estados_uf=estados_uf)
    )


async def fetch(url: str, session: ClientSession):
    """
    Asynchronous get request
    """
    async with session.get(url) as response:
        response_json = await response.json()
        return response_json


async def fetch_many(loop: AbstractEventLoop, urls: List[str]):
    """
    Many asynchronous get requests, gathered
    """
    async with aiohttp.ClientSession() as session:
        tasks = [loop.create_task(fetch(url, session)) for url in urls]
        return await asyncio.gather(*tasks)


async def fetch_single(url: str):
    """
    Single asynchronous get request, gathered
    """
    async with aiohttp.ClientSession() as session:
        return await fetch(url, session)


async def fetch_vacinas_total():
    """
    Single asynchronous get request, gathered
    """
    async with aiohttp.ClientSession(
        auth=aiohttp.BasicAuth(ELASTIC_AUTH[0], ELASTIC_AUTH[1]),
        headers=ELASTIC_HEADERS,
    ) as session:
        return await fetch_elastic(session)


def asnyc_aiohttp_get_all(urls: List[str]):
    """
    Performs asynchronous get requests
    """
    loop = asyncio.get_event_loop()
    return loop.run_until_complete(fetch_many(loop, urls))
