// Copyright 2026 GAIA Contributors
// Auto-generated from JSON Schemas. DO NOT EDIT.

export interface AgentManifest {
    agent_id:              string;
    auth?:                 Auth;
    capabilities:          AgentManifes[];
    endpoint:              string;
    health_endpoint?:      string;
    health_endpoint_spec?: HealthEndpointSpec;
    invoke:                Invoke;
    protocol:              string;
    state_requirements?:   StateRequirements;
    transport:             string;
    version:               string;
}

export interface Auth {
    scopes?:     string[];
    secret_ref?: string;
    type:        string;
}

export interface AgentManifes {
    constraints?:  Constraints;
    description:   string;
    idempotent:    boolean;
    input_schema:  { [key: string]: any };
    name:          string;
    output_schema: { [key: string]: any };
}

export interface Constraints {
    external_io:   boolean;
    mutates_state: boolean;
    read_only:     boolean;
}

export interface HealthEndpointSpec {
    expected_response?: { [key: string]: any };
    method:             string;
    timeout_ms:         number;
}

export interface Invoke {
    async_supported: boolean;
    timeout_ms:      number;
}

export interface StateRequirements {
    max_bytes?: number;
    required:   boolean;
}

export interface AsyncCompletion {
    error?:     Error;
    job_id:     string;
    output?:    any;
    request_id: string;
    success:    boolean;
    type:       string;
}

export interface Error {
    code:      string;
    details?:  { [key: string]: any };
    message:   string;
    retryable: boolean;
}

export interface Event {
    name:               string;
    payload?:           { [key: string]: any };
    previous_event_id?: string;
    sequence_number?:   number;
    step_id?:           string;
    task_id?:           string;
    timestamp:          Date;
    type:               string;
}

export interface Request {
    capability:  string;
    from:        string;
    input:       any;
    mode:        string;
    request_id:  string;
    step_id:     string;
    task_id:     string;
    timeout_ms?: number;
    type:        string;
}

export interface Response {
    error?:     Error;
    job_id?:    string;
    metrics?:   Metrics;
    output?:    any;
    request_id: string;
    success:    boolean;
}

export interface Error {
    code:      string;
    details?:  { [key: string]: any };
    message:   string;
    retryable: boolean;
}

export interface Metrics {
    cost_estimate?: number;
    duration_ms?:   number;
    tokens_used?:   number;
}

export interface Step {
    assigned_agent?:   string;
    async_timeout_ms?: number;
    capability:        string;
    depends_on?:       string[];
    error?:            Error;
    input:             any;
    job_id?:           string;
    output?:           any;
    output_schema?:    { [key: string]: any };
    retry_count:       number;
    status:            string;
    step_id:           string;
}

export interface Error {
    code:      string;
    details?:  { [key: string]: any };
    message:   string;
    retryable: boolean;
}

export interface Task {
    created_at:   Date;
    current_step: number;
    finished_at?: Date;
    goal:         string;
    metadata?:    { [key: string]: any };
    plan?:        Tas[];
    status:       string;
    task_id:      string;
    updated_at:   Date;
}

export interface Tas {
    assigned_agent?:   string;
    async_timeout_ms?: number;
    capability:        string;
    depends_on?:       string[];
    error?:            Error;
    input:             any;
    job_id?:           string;
    output?:           any;
    output_schema?:    { [key: string]: any };
    retry_count:       number;
    status:            string;
    step_id:           string;
}

export interface Error {
    code:      string;
    details?:  { [key: string]: any };
    message:   string;
    retryable: boolean;
}

