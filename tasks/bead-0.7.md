# Bead 0.7: Tool Registration with Metadata

**Phase**: 0
**Priority**: High

## Description
Ensure all tools are registered with proper MCP metadata, especially the `ReadOnlyHint`.

## Tasks
- Use typed `mcp.AddTool` for all tools
- Set `ReadOnlyHint: true` on every tool
- Add clear descriptions and input schemas
- Document tool behavior

## Acceptance Criteria
- All registered tools declare `ReadOnlyHint: true`
- Tool schemas are well-defined
- Descriptions are clear and helpful for agents

## Validation
- Inspect `tools/list` output
- Verify no tool can be mistaken for mutating
- Review against MCP best practices