# Kafka Overview

## Library Selection

| Scenario | Library |
|----------|---------|
| Pure Go, easy local dev, no CGo | `segmentio/kafka-go` |
| Max performance, Confluent ecosystem, transactions | `confluent-kafka-go` |
| Existing Sarama codebase, need mock testing | `IBM/sarama` |
| New project, general purpose | `segmentio/kafka-go` |
| Financial/critical exactly-once requirements | `confluent-kafka-go` |

## Industry Context

- **Netflix**: 2+ trillion msgs/day. One consumer per topic. Invested in message tracing (Inca) early
- **Uber**: Trillions of msgs/day. Built uForwarder (push-based consumer proxy). Partition-level flow control, not all-or-nothing pausing. Federated clusters (~150 nodes each)
- **LinkedIn**: 7+ trillion msgs/day, 100+ clusters, 100k+ topics. Separate topics by category (commands, events, logs, metrics). Partitioning as the scaling primitive

**Key takeaways**: single-responsibility consumers, partition-based parallelism, invest in observability, separate topics by message category.
