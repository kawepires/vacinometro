// To parse this JSON data, do
//
//     final elasticResponse = elasticResponseFromJson(jsonString);

import 'dart:convert';

ElasticResponse elasticResponseFromJson(String str) =>
    ElasticResponse.fromJson(json.decode(str));

String elasticResponseToJson(ElasticResponse data) =>
    json.encode(data.toJson());

class ElasticResponse {
  ElasticResponse({
    required this.took,
    required this.timedOut,
    required this.shards,
    required this.hits,
    required this.aggregations,
  });

  final int took;
  final bool timedOut;
  final Shards shards;
  final Hits hits;
  final Aggregations aggregations;

  factory ElasticResponse.fromJson(Map<String, dynamic> json) =>
      ElasticResponse(
        took: json['took'],
        timedOut: json['timed_out'],
        shards: Shards.fromJson(json['_shards']),
        hits: Hits.fromJson(json['hits']),
        aggregations: Aggregations.fromJson(json['aggregations']),
      );

  Map<String, dynamic> toJson() => {
        'took': took,
        'timed_out': timedOut,
        '_shards': shards.toJson(),
        'hits': hits.toJson(),
        'aggregations': aggregations.toJson(),
      };
}

class Aggregations {
  Aggregations({
    required this.filtros,
  });

  final Filtros filtros;

  factory Aggregations.fromJson(Map<String, dynamic> json) => Aggregations(
        filtros: Filtros.fromJson(json['filtros']),
      );

  Map<String, dynamic> toJson() => {
        'filtros': filtros.toJson(),
      };
}

class Filtros {
  Filtros({
    required this.buckets,
  });

  final Buckets buckets;

  factory Filtros.fromJson(Map<String, dynamic> json) => Filtros(
        buckets: Buckets.fromJson(json['buckets']),
      );

  Map<String, dynamic> toJson() => {
        'buckets': buckets.toJson(),
      };
}

class Buckets {
  Buckets({
    required this.primeiraDose,
    required this.segundaDose,
  });

  final ADose primeiraDose;
  final ADose segundaDose;

  factory Buckets.fromJson(Map<String, dynamic> json) => Buckets(
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
    required this.docCount,
    required this.uniqueDocs,
  });

  final int docCount;
  final UniqueDocs uniqueDocs;

  factory ADose.fromJson(Map<String, dynamic> json) => ADose(
        docCount: json['doc_count'],
        uniqueDocs: UniqueDocs.fromJson(json['unique_docs']),
      );

  Map<String, dynamic> toJson() => {
        'doc_count': docCount,
        'unique_docs': uniqueDocs.toJson(),
      };
}

class UniqueDocs {
  UniqueDocs({
    required this.value,
  });

  final int value;

  factory UniqueDocs.fromJson(Map<String, dynamic> json) => UniqueDocs(
        value: json['value'],
      );

  Map<String, dynamic> toJson() => {
        'value': value,
      };
}

class Hits {
  Hits({
    required this.total,
    required this.maxScore,
    required this.hits,
  });

  final Total total;
  final dynamic maxScore;
  final List<dynamic> hits;

  factory Hits.fromJson(Map<String, dynamic> json) => Hits(
        total: Total.fromJson(json['total']),
        maxScore: json['max_score'],
        hits: List<dynamic>.from(json['hits'].map((x) => x)),
      );

  Map<String, dynamic> toJson() => {
        'total': total.toJson(),
        'max_score': maxScore,
        'hits': List<dynamic>.from(hits.map((x) => x)),
      };
}

class Total {
  Total({
    required this.value,
    required this.relation,
  });

  final int value;
  final String relation;

  factory Total.fromJson(Map<String, dynamic> json) => Total(
        value: json['value'],
        relation: json['relation'],
      );

  Map<String, dynamic> toJson() => {
        'value': value,
        'relation': relation,
      };
}

class Shards {
  Shards({
    required this.total,
    required this.successful,
    required this.skipped,
    required this.failed,
  });

  final int total;
  final int successful;
  final int skipped;
  final int failed;

  factory Shards.fromJson(Map<String, dynamic> json) => Shards(
        total: json['total'],
        successful: json['successful'],
        skipped: json['skipped'],
        failed: json['failed'],
      );

  Map<String, dynamic> toJson() => {
        'total': total,
        'successful': successful,
        'skipped': skipped,
        'failed': failed,
      };
}
