# Bead 0.6: Basic MCP Server Bootstrap

**Phase**: 0
**Priority**: High

## Description
Bootstrap the MCP server supporting both stdio and Streamable HTTP transports.

## Tasks
- Initialize MCP server using official Go SDK
- Support `stdio` transport
- Support `streamable-http` transport
- Implement basic `tools/list` handler
- Add graceful shutdown

## Acceptance Criteria
- Server starts in both stdio and HTTP mode
- `tools/list` returns a valid response
- Server shuts down cleanly

## Validation
- Manual testing with MCP Inspector
- k6 can call the HTTP endpoint
- Server responds correctly to basic MCP requests