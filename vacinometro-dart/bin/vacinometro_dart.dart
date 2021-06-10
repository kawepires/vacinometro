import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'elastic_query.dart';
import 'elastic_response.dart';

// #region Variables, constants and finals

const ibgeProjecoesUrl =
    'https://servicodados.ibge.gov.br/api/v1/projecoes/populacao/';
const elasticsearchUsername = 'imunizacao_public';
const elasticsearchPassword = 'qlto5t&7r_@+#Tlstigi';
const elasticsearchContent = 'application/json';

final elasticsearchUrl =
    Uri.parse('https://imunizacao-es.saude.gov.br/_search');

final elasticsearchAuth = base64Encode(
  utf8.encode('$elasticsearchUsername:$elasticsearchPassword'),
);

final elasticBaseQuery = elasticQueryFromJson(
  jsonDecode(
    File('../../elastic_query.json').readAsStringSync(),
  ),
);

// #endregion Variables, constants and finals

ElasticQuery buildQuery({int? municipioCod, String? estadoUf}) {
  if (municipioCod != null || estadoUf != null) {
    if (municipioCod != null) {
      // elasticBaseQuery.query
    } else {}
  }

  return elasticBaseQuery;
}

// Future<dynamic> makeRequest() async {}

Future<ElasticResponse> makeElasticRequest(ElasticQuery payload) async {
  final client = HttpClient();
  final clientRequest = await client.getUrl(elasticsearchUrl);
  var payloadBytes = utf8.encode(payload.toJson().toString());
  clientRequest.headers.add('Authorization', 'Basic $elasticsearchAuth');
  clientRequest.headers.add('Content-Type', 'application/json');
  clientRequest.headers.add('Content-Length', payloadBytes.length);
  clientRequest.add(payloadBytes);
  final response = await clientRequest.close();

  final contents = await consumeResponse(response);

  final jsonResp = jsonDecode(contents);
  final elasticResponse = elasticResponseFromJson(jsonResp);
  return elasticResponse;
}

Future<String> consumeResponse(HttpClientResponse response) async {
  final contents = StringBuffer();
  await for (var data in response.transform(utf8.decoder)) {
    contents.write(data);
  }
  return contents.toString();
}

Future<void> stageBrasil() async {
  final query = buildQuery();
  final response = await makeElasticRequest(query);
}

Future<void> main(List<String> args) async {}
