# risinggo

Go implementation of RisingWave MCP Server

**Key Characteristics**:
- Strictly read-only (zero mutation capability)
- p99 latency target < 20ms (with 10% variance)
- Powerful generic safe read-only query executor
- Very detailed observability
- Comprehensive Gherkin specification suite
- k6 load testing harness for 200 concurrent users

## Directory Structure

- `design/` - Detailed design document and plan
- `specs/` - Executable Gherkin specifications (BDD)
- `k6-loadtest/` - Load testing harness using k6

See individual folders for details.