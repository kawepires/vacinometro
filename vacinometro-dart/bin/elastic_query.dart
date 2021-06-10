// To parse this JSON data, do
//
//     final elasticQuery = elasticQueryFromJson(jsonString);

import 'dart:convert';

ElasticQuery elasticQueryFromJson(String str) =>
    ElasticQuery.fromJson(json.decode(str));

String elasticQueryToJson(ElasticQuery data) => json.encode(data.toJson());

class ElasticQuery {
  ElasticQuery({
    this.size = 0,
    this.query,
    required this.aggs,
  });

  final int size;
  final ElasticQueryAggs aggs;
  Query? query;

  factory ElasticQuery.fromJson(Map<String, dynamic> json) => ElasticQuery(
        size: json['size'],
        aggs: ElasticQueryAggs.fromJson(json['aggs']),
        query: json['query'] != null ? Query.fromJson(json['query']) : null,
      );

  Map<String, dynamic> toJson() {
    var map = <String, dynamic>{
      'size': size,
      'aggs': aggs.toJson(),
    };
    if (query != null) {
      map['query'] = query!.toJson();
    }
    return map;
  }
}

class Query {
  Query({
    required this.bool,
  });

  final Bool bool;

  factory Query.fromJson(Map<String, dynamic> json) => Query(
        bool: Bool.fromJson(json['bool']),
      );

  Map<String, dynamic> toJson() => {
        'bool': bool.toJson(),
      };
}

class Bool {
  Bool({
    required this.match,
  });

  final QueryMatch match;

  factory Bool.fromJson(Map<String, dynamic> json) => Bool(
        match: QueryMatch.fromJson(json['match']),
      );

  Map<String, dynamic> toJson() => {
        'bool': match.toJson(),
      };
}

class QueryMatch {
  QueryMatch({
    this.estabelecimentoMunicipioCodigo = 0,
    this.estabelecimentoUf = '',
  });

  final int estabelecimentoMunicipioCodigo;
  // Todo: Change to enum.
  final String estabelecimentoUf;

  factory QueryMatch.fromJson(Map<String, dynamic> json) => QueryMatch(
      estabelecimentoUf: json['estabelecimento_uf'],
      estabelecimentoMunicipioCodigo: json['estabelecimento_municipio_codigo']);

  Map<String, dynamic> toJson() {
    var map = <String, dynamic>{};
    if (estabelecimentoMunicipioCodigo > 0) {
      map['estabelecimento_municipio_codigo'] = estabelecimentoMunicipioCodigo;
    } else {
      map['estabelecimento_uf'] = estabelecimentoUf;
    }
    return map;
  }
}

class ElasticQueryAggs {
  ElasticQueryAggs({
    required this.filtros,
  });

  final Filtros filtros;

  factory ElasticQueryAggs.fromJson(Map<String, dynamic> json) =>
      ElasticQueryAggs(
        filtros: Filtros.fromJson(json['filtros']),
      );

  Map<String, dynamic> toJson() => {
        'filtros': filtros.toJson(),
      };
}

class Filtros {
  Filtros({
    required this.filters,
    required this.aggs,
  });

  final FiltrosFilters filters;
  final FiltrosAggs aggs;

  factory Filtros.fromJson(Map<String, dynamic> json) => Filtros(
        filters: FiltrosFilters.fromJson(json['filters']),
        aggs: FiltrosAggs.fromJson(json['aggs']),
      );

  Map<String, dynamic> toJson() => {
        'filters': filters.toJson(),
        'aggs': aggs.toJson(),
      };
}

class FiltrosAggs {
  FiltrosAggs({
    required this.uniqueDocs,
  });

  final UniqueDocs uniqueDocs;

  factory FiltrosAggs.fromJson(Map<String, dynamic> json) => FiltrosAggs(
        uniqueDocs: UniqueDocs.fromJson(json['unique_docs']),
      );

  Map<String, dynamic> toJson() => {
        'unique_docs': uniqueDocs.toJson(),
      };
}

class UniqueDocs {
  UniqueDocs({
    required this.cardinality,
  });

  final Cardinality cardinality;

  factory UniqueDocs.fromJson(Map<String, dynamic> json) => UniqueDocs(
        cardinality: Cardinality.fromJson(json['cardinality']),
      );

  Map<String, dynamic> toJson() => {
        'cardinality': cardinality.toJson(),
      };
}

class Cardinality {
  Cardinality({
    required this.field,
  });

  final String field;

  factory Cardinality.fromJson(Map<String, dynamic> json) => Cardinality(
        field: json['field'],
      );

  Map<String, dynamic> toJson() => {
        'field': field,
      };
}

class FiltrosFilters {
  FiltrosFilters({
    required this.filters,
  });

  final FiltersFilters filters;

  factory FiltrosFilters.fromJson(Map<String, dynamic> json) => FiltrosFilters(
        filters: FiltersFilters.fromJson(json['filters']),
      );

  Map<String, dynamic> toJson() => {
        'filters': filters.toJson(),
      };
}

class FiltersFilters {
  FiltersFilters({
    required this.primeiraDose,
    required this.segundaDose,
  });

  final ADose primeiraDose;
  final ADose segundaDose;

  factory FiltersFilters.fromJson(Map<String, dynamic> json) => FiltersFilters(
        primeiraDose: ADose.fromJson(json['primeira_dose']),
        segundaDose: ADose.fromJson(json['segunda_dose']),
      );

  Map<String, dynamic> toJson() => {
        'primeira_dose': primeiraDose.toJson(),
        'segunda_dose': segundaDose.toJson(),
      };
}

class ADose {
  ADose({
    required this.match,
  });

  final Match match;

  factory ADose.fromJson(Map<String, dynamic> json) => ADose(
        match: Match.fromJson(json['match']),
      );

  Map<String, dynamic> toJson() => {
        'match': match.toJson(),
      };
}

class Match {
  Match({
    required this.vacinaDescricaoDose,
  });

  final VacinaDescricaoDose vacinaDescricaoDose;

  factory Match.fromJson(Map<String, dynamic> json) => Match(
        vacinaDescricaoDose:
            VacinaDescricaoDose.fromJson(json['vacina_descricao_dose']),
      );

  Map<String, dynamic> toJson() => {
        'vacina_descricao_dose': vacinaDescricaoDose.toJson(),
      };
}

class VacinaDescricaoDose {
  VacinaDescricaoDose({
    required this.query,
    required this.vacinaDescricaoDoseOperator,
  });

  final String query;
  final String vacinaDescricaoDoseOperator;

  factory VacinaDescricaoDose.fromJson(Map<String, dynamic> json) =>
      VacinaDescricaoDose(
        query: json['query'],
        vacinaDescricaoDoseOperator: json['operator'],
      );

  Map<String, dynamic> toJson() => {
        'query': query,
        'operator': vacinaDescricaoDoseOperator,
      };
}
