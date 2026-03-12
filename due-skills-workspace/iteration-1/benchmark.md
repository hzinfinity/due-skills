# due-skills 评估报告 (Iteration 1)

## 汇总结果

| 配置 | 通过率 | 通过数 | 总断言数 |
|------|--------|--------|----------|
| With Skill | 100.0% | 42 | 42 |
| Without Skill | 45.2% | 19 | 42 |
| **提升** | **+54.8%** | - | - |

## 各评估用例详情

### websocket-gate-server

| 配置 | 通过率 | 通过数 |
|------|--------|--------|
| With Skill | 100.0% | 7/7 |
| Without Skill | 0.0% | 0/7 |
| **提升** | **+100.0%** | - |

**断言详情:**

| 断言 | With Skill | Without Skill |
|------|------------|---------------|
| uses_container_pattern | ✓ | ✗ |
| uses_gate_component | ✓ | ✗ |
| uses_websocket_server | ✓ | ✗ |
| configures_redis_locator | ✓ | ✗ |
| configures_consul_registry | ✓ | ✗ |
| sets_max_connections | ✓ | ✗ |
| uses_v2_imports | ✓ | ✗ |

### node-router-handler

| 配置 | 通过率 | 通过数 |
|------|--------|--------|
| With Skill | 100.0% | 9/9 |
| Without Skill | 88.9% | 8/9 |
| **提升** | **+11.1%** | - |

**断言详情:**

| 断言 | With Skill | Without Skill |
|------|------------|---------------|
| uses_container_pattern | ✓ | ✓ |
| uses_node_component | ✓ | ✓ |
| registers_route_handler | ✓ | ✓ |
| uses_node_context | ✓ | ✓ |
| parses_request | ✓ | ✓ |
| sends_response | ✓ | ✓ |
| login_route_equals_1 | ✓ | ✗ |
| request_has_uid_and_token | ✓ | ✓ |
| response_has_code_and_message | ✓ | ✓ |

### mesh-microservice-setup

| 配置 | 通过率 | 通过数 |
|------|--------|--------|
| With Skill | 100.0% | 8/8 |
| Without Skill | 37.5% | 3/8 |
| **提升** | **+62.5%** | - |

**断言详情:**

| 断言 | With Skill | Without Skill |
|------|------------|---------------|
| uses_container_pattern | ✓ | ✗ |
| uses_mesh_component | ✓ | ✗ |
| uses_rpcx_transporter | ✓ | ✗ |
| registers_service_provider | ✓ | ✗ |
| implements_getuser_method | ✓ | ✓ |
| getuser_request_has_uid | ✓ | ✓ |
| uses_context_context | ✓ | ✓ |
| configures_redis_locator | ✓ | ✗ |

### chat-room-complete

| 配置 | 通过率 | 通过数 |
|------|--------|--------|
| With Skill | 100.0% | 10/10 |
| Without Skill | 0.0% | 0/10 |
| **提升** | **+100.0%** | - |

**断言详情:**

| 断言 | With Skill | Without Skill |
|------|------------|---------------|
| has_gate_service | ✓ | ✗ |
| has_node_service | ✓ | ✗ |
| websocket_port_8800 | ✓ | ✗ |
| has_login_handler | ✓ | ✗ |
| has_chat_handler | ✓ | ✗ |
| has_logout_handler | ✓ | ✗ |
| uses_message_broadcast | ✓ | ✗ |
| uses_redis_locator | ✓ | ✗ |
| uses_consul_registry | ✓ | ✗ |
| three_routes_defined | ✓ | ✗ |

### redis-eventbus-config

| 配置 | 通过率 | 通过数 |
|------|--------|--------|
| With Skill | 100.0% | 8/8 |
| Without Skill | 100.0% | 8/8 |
| **提升** | **+0.0%** | - |

**断言详情:**

| 断言 | With Skill | Without Skill |
|------|------------|---------------|
| creates_redis_client | ✓ | ✓ |
| creates_eventbus | ✓ | ✓ |
| publishes_user_login | ✓ | ✓ |
| subscribes_to_event | ✓ | ✓ |
| calls_publish | ✓ | ✓ |
| uses_v2_eventbus_import | ✓ | ✓ |
| has_event_struct | ✓ | ✓ |
| uses_json_marshal | ✓ | ✓ |

