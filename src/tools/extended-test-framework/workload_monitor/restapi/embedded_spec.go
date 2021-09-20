// Code generated by go-swagger; DO NOT EDIT.

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

var (
	// SwaggerJSON embedded version of the swagger document used at generation time
	SwaggerJSON json.RawMessage
	// FlatSwaggerJSON embedded flattened version of the swagger document used at generation time
	FlatSwaggerJSON json.RawMessage
)

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "description": "MayaData System Test Framework API",
    "title": "Test Framework API",
    "version": "1.0.1"
  },
  "basePath": "/api/v1",
  "paths": {
    "/wm/registrants/{rid}": {
      "get": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "returns all workloads registered by the specified registrant",
        "operationId": "GetWorkloadsByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "corresponding Workload item(s) returned to caller",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Workload"
              }
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "no corresponding workload(s) found"
          }
        }
      },
      "delete": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "deletes all workloads registered by the specified registrant",
        "operationId": "DeleteWorkloadsByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Workload(s) deleted",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "resource not found"
          }
        }
      }
    },
    "/wm/registrants/{rid}/workloads/{wid}": {
      "get": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "returns a specific workload registration (identified  workload and registrant)",
        "operationId": "GetWorkloadByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          },
          {
            "type": "string",
            "format": "uuid",
            "description": "workload uid",
            "name": "wid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "corresponding Workload item returned to caller",
            "schema": {
              "$ref": "#/definitions/Workload"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "no corresponding workload found"
          }
        }
      },
      "put": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "Creates/Updates a workload registration for a specific registratant",
        "operationId": "PutWorkloadByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          },
          {
            "type": "string",
            "format": "uuid",
            "description": "workload uid",
            "name": "wid",
            "in": "path",
            "required": true
          },
          {
            "name": "body",
            "in": "body",
            "schema": {
              "$ref": "#/definitions/WorkloadSpec"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Workload registered/updated",
            "schema": {
              "$ref": "#/definitions/Workload"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          }
        }
      },
      "delete": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "deletes a specific workload registration (identified  workload and registrant)",
        "operationId": "DeleteWorkloadByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          },
          {
            "type": "string",
            "format": "uuid",
            "description": "workload uid",
            "name": "wid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Workload deleted",
            "schema": {
              "$ref": "#/definitions/Workload"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "resource not found"
          }
        }
      }
    },
    "/wm/workloads": {
      "get": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "returns all workloads registered with the monitor",
        "operationId": "GetWorkloads",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "metadata.uid of Pod which registered the workload",
            "name": "registrantId",
            "in": "query"
          },
          {
            "type": "string",
            "description": "workload (pod) name",
            "name": "name",
            "in": "query"
          },
          {
            "type": "string",
            "description": "workload (pod) namespace",
            "name": "namespace",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "Workload items were returned",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Workload"
              }
            }
          },
          "404": {
            "description": "no corresponding workload(s) were returned"
          }
        }
      }
    }
  },
  "definitions": {
    "Event": {
      "allOf": [
        {
          "type": "object",
          "properties": {
            "id": {
              "type": "string",
              "format": "uuid"
            },
            "loggedDateTime": {
              "type": "string",
              "format": "date-time"
            }
          }
        },
        {
          "$ref": "#/definitions/EventSpec"
        }
      ]
    },
    "EventClassEnum": {
      "type": "string",
      "enum": [
        "FAIL",
        "INFO",
        "WARN"
      ]
    },
    "EventSourceClassEnum": {
      "type": "string",
      "enum": [
        "workload-monitor",
        "log-monitor",
        "resouce-monitor"
      ]
    },
    "EventSpec": {
      "type": "object",
      "required": [
        "sourceClass",
        "sourceInstance",
        "class",
        "message"
      ],
      "properties": {
        "class": {
          "$ref": "#/definitions/EventClassEnum"
        },
        "data": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "message": {
          "type": "string"
        },
        "resource": {
          "type": "string"
        },
        "sourceClass": {
          "$ref": "#/definitions/EventSourceClassEnum"
        },
        "sourceInstance": {
          "type": "string"
        }
      }
    },
    "RFC1123Label": {
      "description": "kubernetes label name conforming to RFC1123",
      "type": "string",
      "pattern": "^[a-z0-9][a-z0-9\\-]{0,61}[a-z0-9]{1}"
    },
    "RequestOutcome": {
      "type": "object",
      "properties": {
        "details": {
          "type": "string",
          "example": "reason(s) why the request cannot be handled"
        },
        "items_affected": {
          "description": "number of items affected (e.g.) by the request",
          "type": "integer",
          "format": "int64"
        },
        "result": {
          "type": "string",
          "enum": [
            "REFUSED",
            "OK"
          ]
        }
      }
    },
    "Workload": {
      "allOf": [
        {
          "type": "object",
          "properties": {
            "id": {
              "description": "metadata.uid of workload pod",
              "type": "string",
              "format": "uuid"
            },
            "name": {
              "$ref": "#/definitions/RFC1123Label"
            },
            "namespace": {
              "$ref": "#/definitions/RFC1123Label"
            }
          }
        },
        {
          "$ref": "#/definitions/WorkloadSpec"
        }
      ]
    },
    "WorkloadSpec": {
      "type": "object",
      "required": [
        "violations"
      ],
      "properties": {
        "violations": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/WorkloadViolationEnum"
          }
        }
      }
    },
    "WorkloadViolationEnum": {
      "type": "string",
      "enum": [
        "RESTARTED",
        "TERMINATED",
        "NOT_PRESENT"
      ]
    }
  },
  "parameters": {
    "EventClassEnumQueryParam": {
      "enum": [
        "FAIL",
        "INFO",
        "WARN"
      ],
      "type": "string",
      "description": "event class",
      "name": "class",
      "in": "query"
    },
    "EventSourceClassEnumQueryParam": {
      "enum": [
        "workload-monitor",
        "log-monitor",
        "resouce-monitor"
      ],
      "type": "string",
      "description": "source class",
      "name": "sourceClass",
      "in": "query"
    },
    "JiraKeyPathParam": {
      "pattern": "^[A-Z]{2,3}-\\d{1,4}$",
      "type": "string",
      "format": "Jira issue key",
      "description": "Test Plan Id",
      "name": "id",
      "in": "path",
      "required": true
    }
  },
  "tags": [
    {
      "description": "Workload Monitor",
      "name": "workload-monitor"
    }
  ]
}`))
	FlatSwaggerJSON = json.RawMessage([]byte(`{
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "description": "MayaData System Test Framework API",
    "title": "Test Framework API",
    "version": "1.0.1"
  },
  "basePath": "/api/v1",
  "paths": {
    "/wm/registrants/{rid}": {
      "get": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "returns all workloads registered by the specified registrant",
        "operationId": "GetWorkloadsByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "corresponding Workload item(s) returned to caller",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Workload"
              }
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "no corresponding workload(s) found"
          }
        }
      },
      "delete": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "deletes all workloads registered by the specified registrant",
        "operationId": "DeleteWorkloadsByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Workload(s) deleted",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "resource not found"
          }
        }
      }
    },
    "/wm/registrants/{rid}/workloads/{wid}": {
      "get": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "returns a specific workload registration (identified  workload and registrant)",
        "operationId": "GetWorkloadByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          },
          {
            "type": "string",
            "format": "uuid",
            "description": "workload uid",
            "name": "wid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "corresponding Workload item returned to caller",
            "schema": {
              "$ref": "#/definitions/Workload"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "no corresponding workload found"
          }
        }
      },
      "put": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "Creates/Updates a workload registration for a specific registratant",
        "operationId": "PutWorkloadByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          },
          {
            "type": "string",
            "format": "uuid",
            "description": "workload uid",
            "name": "wid",
            "in": "path",
            "required": true
          },
          {
            "name": "body",
            "in": "body",
            "schema": {
              "$ref": "#/definitions/WorkloadSpec"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Workload registered/updated",
            "schema": {
              "$ref": "#/definitions/Workload"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          }
        }
      },
      "delete": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "deletes a specific workload registration (identified  workload and registrant)",
        "operationId": "DeleteWorkloadByRegistrant",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "registrant uid",
            "name": "rid",
            "in": "path",
            "required": true
          },
          {
            "type": "string",
            "format": "uuid",
            "description": "workload uid",
            "name": "wid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Workload deleted",
            "schema": {
              "$ref": "#/definitions/Workload"
            }
          },
          "400": {
            "description": "Bad request (malformed/invalid resource id)",
            "schema": {
              "$ref": "#/definitions/RequestOutcome"
            }
          },
          "404": {
            "description": "resource not found"
          }
        }
      }
    },
    "/wm/workloads": {
      "get": {
        "tags": [
          "workload-monitor"
        ],
        "summary": "returns all workloads registered with the monitor",
        "operationId": "GetWorkloads",
        "parameters": [
          {
            "type": "string",
            "format": "uuid",
            "description": "metadata.uid of Pod which registered the workload",
            "name": "registrantId",
            "in": "query"
          },
          {
            "type": "string",
            "description": "workload (pod) name",
            "name": "name",
            "in": "query"
          },
          {
            "type": "string",
            "description": "workload (pod) namespace",
            "name": "namespace",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "Workload items were returned",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Workload"
              }
            }
          },
          "404": {
            "description": "no corresponding workload(s) were returned"
          }
        }
      }
    }
  },
  "definitions": {
    "Event": {
      "allOf": [
        {
          "type": "object",
          "properties": {
            "id": {
              "type": "string",
              "format": "uuid"
            },
            "loggedDateTime": {
              "type": "string",
              "format": "date-time"
            }
          }
        },
        {
          "$ref": "#/definitions/EventSpec"
        }
      ]
    },
    "EventClassEnum": {
      "type": "string",
      "enum": [
        "FAIL",
        "INFO",
        "WARN"
      ]
    },
    "EventSourceClassEnum": {
      "type": "string",
      "enum": [
        "workload-monitor",
        "log-monitor",
        "resouce-monitor"
      ]
    },
    "EventSpec": {
      "type": "object",
      "required": [
        "sourceClass",
        "sourceInstance",
        "class",
        "message"
      ],
      "properties": {
        "class": {
          "$ref": "#/definitions/EventClassEnum"
        },
        "data": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "message": {
          "type": "string"
        },
        "resource": {
          "type": "string"
        },
        "sourceClass": {
          "$ref": "#/definitions/EventSourceClassEnum"
        },
        "sourceInstance": {
          "type": "string"
        }
      }
    },
    "RFC1123Label": {
      "description": "kubernetes label name conforming to RFC1123",
      "type": "string",
      "pattern": "^[a-z0-9][a-z0-9\\-]{0,61}[a-z0-9]{1}"
    },
    "RequestOutcome": {
      "type": "object",
      "properties": {
        "details": {
          "type": "string",
          "example": "reason(s) why the request cannot be handled"
        },
        "items_affected": {
          "description": "number of items affected (e.g.) by the request",
          "type": "integer",
          "format": "int64",
          "minimum": 0
        },
        "result": {
          "type": "string",
          "enum": [
            "REFUSED",
            "OK"
          ]
        }
      }
    },
    "Workload": {
      "allOf": [
        {
          "type": "object",
          "properties": {
            "id": {
              "description": "metadata.uid of workload pod",
              "type": "string",
              "format": "uuid"
            },
            "name": {
              "$ref": "#/definitions/RFC1123Label"
            },
            "namespace": {
              "$ref": "#/definitions/RFC1123Label"
            }
          }
        },
        {
          "$ref": "#/definitions/WorkloadSpec"
        }
      ]
    },
    "WorkloadSpec": {
      "type": "object",
      "required": [
        "violations"
      ],
      "properties": {
        "violations": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/WorkloadViolationEnum"
          }
        }
      }
    },
    "WorkloadViolationEnum": {
      "type": "string",
      "enum": [
        "RESTARTED",
        "TERMINATED",
        "NOT_PRESENT"
      ]
    }
  },
  "parameters": {
    "EventClassEnumQueryParam": {
      "enum": [
        "FAIL",
        "INFO",
        "WARN"
      ],
      "type": "string",
      "description": "event class",
      "name": "class",
      "in": "query"
    },
    "EventSourceClassEnumQueryParam": {
      "enum": [
        "workload-monitor",
        "log-monitor",
        "resouce-monitor"
      ],
      "type": "string",
      "description": "source class",
      "name": "sourceClass",
      "in": "query"
    },
    "JiraKeyPathParam": {
      "pattern": "^[A-Z]{2,3}-\\d{1,4}$",
      "type": "string",
      "format": "Jira issue key",
      "description": "Test Plan Id",
      "name": "id",
      "in": "path",
      "required": true
    }
  },
  "tags": [
    {
      "description": "Workload Monitor",
      "name": "workload-monitor"
    }
  ]
}`))
}
