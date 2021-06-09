# Vacinômetro - Agregador de dados do COVID-19 no Brasil

O propósito desse repositório é agregar os dados sobre a vacinação, em **várias linguagens de programação**, do [OPEN DATASUS](https://opendatasus.saude.gov.br/dataset/covid-19-vacinacao) utilizando a API pública do Elasticsearch, providenciada por eles.

## Detalhes do projeto

### Estimativa populacional
Para Brasil e Estados, os dados são recuperados da [API de projeções](https://servicodados.ibge.gov.br/api/docs/projecoes) do IBGE.

Para municípios, os dados são recuperados do próprio site do IBGE, na seção de [estimativas populacionais](https://www.ibge.gov.br/estatisticas/sociais/populacao/9103-estimativas-de-populacao.html?=&t=o-que-e), até então dos dados mais recentes, sendos estes de 2020.

A API de projeções não fornece dados populacionais para municípios.

### Elasticsearch
Os dados da campanha da vacinação não são 100% confiáveis. Alguns atritos:

#### Cógidos de identificação de municípios do IBGE
Em todos os casos, temos que o parâmetro de pesquisa para municípios (`estabelecimento_municipio_codigo`) do Elasticsearch, espera que o identificador seja enviado sem o último dígito. Exemplo:

1. Manicoré, código: `1302702`;
2. Codigo fonte intermediador remove o número "2";
3. Chamada pra API do Elasticsearch é invocada: `estabelecimento_municipio_codigo: 130270`;
4. Retorno da API com os dados da vacinação.

Ainda não encotrei o motivo de funcionar assim.

#### Identificação da dose
As doses vêm ou com espaços nas laterais, como "    1ª Dose" ou "    2ª Dose", e quando a dose não é especificada: "    Dose ", sem um padrão, pois às vezes temos dados "organizados", sem espaços e com textos exatos.

O que torna uma pouco difícil procura por `.keyword` que além de aceletar a query, procura por valores exatos (*exact match*).

#### Pareamento com outras fontes
Fontes como:

- https://qsprod.saude.gov.br/extensions/DEMAS_C19Vacina/DEMAS_C19Vacina.html
- https://especiais.g1.globo.com/bemestar/vacina/2021/mapa-brasil-vacina-covid/
- https://datastudio.google.com/reporting/f82be37d-7f5f-441c-b96c-464033259509/page/dai8B
- http://www.giscard.com.br/coronavirus/vacinometro-covid19-brasil.php

Têm valores diferentes, dificultando a comparação de dados. Felizmente, esse script se aproxima dos valores de https://qsprod.saude.gov.br/extensions/DEMAS_C19Vacina/DEMAS_C19Vacina.html site oficial de tracking do governo.