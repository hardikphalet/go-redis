# Technical Requirements Document (TRD)

**Project Title**: Redis Clone (Subset Implementation)  
**Author**: Hardik Phalet  
**Date**: 4 June 2025  
**Language**: Go (Golang)

---

## 1. Objective

The goal of this project is to implement a simplified version of the Redis key-value store, focusing on core functionality. This implementation should include both a server and a client that communicate using the Redis Serialization Protocol (RESP).

---

## 2. System Overview

The system will consist of two primary components:

- **In-Memory Database Server**:  
  A lightweight, single-instance server responsible for handling client requests and managing in-memory data structures.

- **Client**:  
  A command-line client that sends requests to the server using the RESP protocol and prints the server's responses.

---

## 3. Communication Protocol

- **Protocol**: RESP (Redis Serialization Protocol)  
- **Reference**: [RESP Specification](https://redis.io/docs/reference/protocol-spec/)  
- All communication between the client and server must adhere strictly to the RESP specification.

---

## 4. Supported Commands

The following Redis commands must be implemented with full parity to their latest documented behavior:

| Command  | Description                                 |
|----------|---------------------------------------------|
| `SET`    | Set the value of a key                      |
| `GET`    | Get the value of a key                      |
| `DEL`    | Delete one or more keys                     |
| `EXPIRE` | Set a timeout on a key                      |
| `TTL`    | Get the remaining time to live of a key     |
| `KEYS`   | Find all keys matching a given pattern      |
| `ZADD`   | Add one or more members to a sorted set     |
| `ZRANGE` | Return a range of members in a sorted set   |

Refer to the [Redis Command Reference](https://redis.io/commands/) for detailed semantics of each.

---

## 5. Functional Requirements

### 5.1 Server

- Accepts multiple client connections over TCP.
- Parses and processes RESP-formatted requests.
- Executes supported Redis commands.
- Manages an in-memory data store.
- Supports key expiration (`EXPIRE`, `TTL`).

### 5.2 Client

- Connects to the server over TCP.
- Encodes commands into RESP format.
- Sends commands and displays responses.

---

## 6. Non-Functional Requirements

- **Correctness**: All commands must behave as per the latest Redis documentation.
- **Simplicity**: Code should be clean, idiomatic Go, with minimal dependencies.
- **Testability**: Code must be testable with a minimal test suite for core commands.
- **Performance**: Performance optimizations are not required but appreciated.
- **Time**: Total time taken will be considered in evaluation.

---

## 7. Evaluation Criteria

1. **Correctness** – Commands behave exactly like the latest Redis version.
2. **Simplicity** – Code is easy to read and understand.
3. **Testability** – There are sufficient and meaningful tests.
4. **Time Taken** – Reasonable time to complete the assignment.

---

## 8. Deliverables

- Source code for the server and client (written in Go).
- Instructions to build and run the project.
- Basic test cases (unit or integration).
- (Optional) A `README.md` file with architecture notes or trade-offs made.

---

## 9. Notes

- No backward compatibility with older Redis versions is required.
- No need to persist data to disk – this is an **in-memory** implementation.
- Use only standard Go packages unless a third-party package is essential.

# Dev logs

- Taking inspiration from the Redis, first step for me would be to implement a multi-threaded TCP server. 
- Structure of this project is going to be simple. We have a server, handler, commands, store and a resp translation unit
- In-memory store is easy to implement, but concurrency of writes and reads might pose an issue
- Go provides a read-write mutex, might want to use it
- Graceful shutdown also becomes an issue, since we are accepting requests on different goroutines
- Go provides a WaitGroup, so will use it to implement graceful shutdown
- Only major part that remain is: 
  1. Writer and reader for RESP protocol
  2. Parsing and routing in handler
  3. adding other redis commands 