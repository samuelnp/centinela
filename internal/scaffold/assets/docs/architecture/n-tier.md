# N-Tier / Layered Architecture Guide

> This document describes the **N-Tier (Layered) archetype**. It is one of five supported architecture archetypes in Centinela. See [architecture-overview.md](architecture-overview.md) to confirm this is the right pattern for your project before reading further. Examples use Express.js (TypeScript), FastAPI (Python), and Go's `net/http` to show how the pattern applies across languages.

---

## When to Choose N-Tier

Choose N-Tier when:

- You are building a REST API, GraphQL server, or microservice
- The request-response lifecycle is the primary interaction model
- Business logic is moderate — orchestration and transformation rather than a rich domain
- You need a pattern that is universally understood across teams and languages

Do NOT choose N-Tier when:

- Domain logic is complex enough to justify explicit domain modelling (use Hexagonal)
- You are on Rails/Django/Laravel (use Rails-native — do not fight the framework)
- You are building a game or simulation (use ECS)

---

## Core Idea

Requests flow in one direction through three fixed layers. Each layer has a single responsibility. No layer skips another.

```
Request → Handler → Service → Repository → Database / External API
Response ←         ←          ←
```

The Handler parses and validates. The Service contains business logic. The Repository accesses data. That is the entire pattern. Its power comes from strict enforcement — the moment a handler queries a database directly, the architecture degrades.

---

## Layer Details

### Handler / Controller Layer

The entry point for every request. Handles HTTP concerns and nothing else.

**What it contains:**

- Route definitions (or routing decorators)
- Request parsing — extract path params, query params, body, headers
- Input validation — reject malformed requests before they reach the service
- Calling one service method
- Formatting and returning the HTTP response

```
# Express.js (TypeScript)
src/handlers/
  posts/
    createPostHandler.ts
    getPostHandler.ts
    listPostsHandler.ts
    deletePostHandler.ts
  users/
    registerUserHandler.ts
    getUserHandler.ts
  comments/
    addCommentHandler.ts
  router.ts                    # mounts all route handlers

# FastAPI (Python)
app/routers/
  posts.py                     # @router.get("/posts"), @router.post("/posts"), etc.
  users.py
  comments.py
  __init__.py

# Go (net/http)
internal/handler/
  post_handler.go
  user_handler.go
  comment_handler.go
  routes.go                    # mux.HandleFunc("/posts", ...)
```

**Rules:**

1. A handler contains zero business logic. Its only conditional statements are input validation checks (`if body.title == ""`) and error-to-HTTP-status mappings.
2. A handler calls exactly one service method. If a handler calls two service methods and combines their results, that orchestration belongs in the service.
3. No database imports in handlers. No ORM, no SQL, no query builders.
4. Request validation is synchronous and pure — it checks shape and types, not business constraints ("title is required" yes; "title must be unique" no — that is a business rule for the service).
5. Response formatting (selecting which fields to include, renaming fields, adding pagination envelopes) belongs here, not in the service.
6. Authentication middleware (parsing JWT, checking session) is separate from handlers — use framework middleware.
7. Handler files contain one handler function or class per HTTP operation. One file per resource group.

**Express.js example:**

```typescript
// src/handlers/posts/createPostHandler.ts
export async function createPostHandler(req: Request, res: Response): Promise<void> {
  const { title, body, authorId } = req.body;
  if (!title || !body || !authorId) {
    res.status(400).json({ error: "title, body, and authorId are required" });
    return;
  }
  const result = await postService.createPost({ title, body, authorId });
  res.status(201).json({ id: result.id, title: result.title });
}
```

**FastAPI example:**

```python
# app/routers/posts.py
@router.post("/posts", status_code=201)
async def create_post(payload: CreatePostRequest, service: PostService = Depends(get_post_service)):
    result = await service.create_post(payload.title, payload.body, payload.author_id)
    return {"id": result.id, "title": result.title}
```

**Go example:**

```go
// internal/handler/post_handler.go
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreatePostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    result, err := h.service.CreatePost(r.Context(), req.Title, req.Body, req.AuthorID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}
```

---

### Service / Business Layer

Where business logic lives. The service is the application's brain — it makes decisions, enforces constraints, and orchestrates data access.

**What it contains:**

- Business rules and constraints ("a post can only be published if the author has a verified email")
- Orchestration of multiple repository calls
- Data transformation (computing derived values, aggregating data)
- Calling external services (payment processors, email senders) — via interfaces/ports, not direct imports

```
# Express.js (TypeScript)
src/services/
  PostService.ts               # createPost, publishPost, archivePost, listPublishedPosts
  UserService.ts               # registerUser, deactivateUser, changeEmail
  CommentService.ts            # addComment, flagComment
  NotificationService.ts       # wraps email/push notification client

# FastAPI (Python)
app/services/
  post_service.py
  user_service.py
  comment_service.py
  notification_service.py

# Go
internal/service/
  post_service.go
  user_service.go
  comment_service.go
  notification_service.go
```

**Rules:**

1. Services have no imports from handler packages and no HTTP concepts (no `Request`, no `Response`, no status codes).
2. Services receive interfaces/protocols for their dependencies (repository, external clients) via constructor injection — not concrete implementations.
3. Business rule violations are returned as typed errors or result types — not HTTP errors.
4. A service method corresponds to one business operation. It is permissible for a service method to call multiple repository methods.
5. Services must not call other service classes directly in a way that creates circular dependencies. If two services need each other's logic, extract shared logic to a helper or reconsider the boundaries.
6. Data returned from service methods is typed (typed interfaces, Pydantic models, Go structs) — no raw `any` or untyped dicts.
7. Services are the correct location for transactional boundaries — if two repository writes must succeed or fail together, the service coordinates that transaction.

**TypeScript example:**

```typescript
// src/services/PostService.ts
export class PostService {
  constructor(
    private readonly postRepo: PostRepository,
    private readonly userRepo: UserRepository,
    private readonly notifier: NotificationService,
  ) {}

  async createPost(input: CreatePostInput): Promise<PostRecord> {
    const author = await this.userRepo.findById(input.authorId);
    if (!author) throw new NotFoundError(`User ${input.authorId} not found`);
    if (!author.emailVerified) throw new ForbiddenError("Author must verify email before posting");

    const post = await this.postRepo.save({
      title: input.title,
      body: input.body,
      authorId: input.authorId,
      status: "draft",
    });

    await this.notifier.sendDraftCreatedAlert(author.email, post.id);
    return post;
  }
}
```

---

### Repository / Data Layer

Handles all data access. The repository is the only layer that knows about databases, query languages, or external data APIs.

**What it contains:**

- Database queries (ORM calls, raw SQL, query builders)
- External API calls that retrieve or persist data (third-party services treated as data sources)
- Caching logic (read-through, write-through)
- Data mapping between database representations and the types the service layer uses

```
# Express.js (TypeScript)
src/repositories/
  PostRepository.ts            # findById, findByAuthor, save, delete, listPublished
  UserRepository.ts            # findById, findByEmail, save, updateEmailVerified
  CommentRepository.ts         # findByPost, save, delete

# FastAPI (Python)
app/repositories/
  post_repository.py
  user_repository.py
  comment_repository.py

# Go
internal/repository/
  post_repository.go
  user_repository.go
  comment_repository.go
```

**Rules:**

1. Repository methods have no business logic. `FindPublishedPostsByAuthor` is a filter — it executes a query. Deciding whether publishing is allowed is a business rule for the service.
2. Repository method names describe data operations, not business operations. `Save`, `FindById`, `Delete`, `ListWhere` — not `PublishPost` (that implies a business state change).
3. Repositories return the service layer's data types, not raw DB rows or HTTP response objects. A mapper within the repository file handles the translation.
4. Repositories implement an interface so services can receive a fake/mock in tests.
5. Database-specific code (transactions, connection pools, query hints) is contained within the repository — never leaks upward.
6. No HTTP concepts in repositories. If a repository calls an external API, it translates the response into the application's data types before returning.
7. Repositories do not call services.

**TypeScript example:**

```typescript
// src/repositories/PostRepository.ts
export interface PostRepository {
  findById(id: string): Promise<PostRecord | null>;
  findByAuthor(authorId: string): Promise<PostRecord[]>;
  save(data: NewPostData): Promise<PostRecord>;
  delete(id: string): Promise<void>;
}

export class PostgresPostRepository implements PostRepository {
  constructor(private readonly db: Knex) {}

  async findById(id: string): Promise<PostRecord | null> {
    const row = await this.db("posts").where({ id }).first();
    return row ? this.toRecord(row) : null;
  }

  private toRecord(row: PostRow): PostRecord {
    return { id: row.id, title: row.title, body: row.body, authorId: row.author_id, status: row.status };
  }
}
```

---

## Dependency Direction

```
Handler → Service → Repository
```

This is the only permitted direction. Any import that reverses or skips this chain is a violation.

- Handler imports Service (by interface)
- Service imports Repository (by interface)
- Repository imports nothing in the application (only DB drivers and external SDKs)
- No layer imports the layer above it

---

## Forbidden Patterns (G2)

| Pattern | Violation | Where it belongs |
|---|---|---|
| DB query in a handler | Handler imports Repository or ORM directly | Service → Repository |
| HTTP status code in a service | Service knows about `res.status(404)` | Handler maps errors to status codes |
| Business rule in a repository | `findPublishablePostsByAuthor` filters on business eligibility | Service decides what is publishable |
| Service calling another service via concrete class | `new CommentService()` inside `PostService` | Inject via interface, or reconsider the boundary |
| Repository calling a service | Data layer calling business logic | Reverse this: service calls repository |
| Handler calling two services and merging results | Orchestration logic in the handler | Extract a service method that does the orchestration |
| Raw `any` types crossing layer boundaries | Untyped data propagation | Define typed interfaces for all layer boundaries |

---

## What "No Business Logic in Outer Layer" Means (G7)

In N-Tier, the "outer layer" is the **Handler / Controller layer**.

**Violations:**

```typescript
// BAD: handler contains a business rule
export async function publishPostHandler(req: Request, res: Response): Promise<void> {
  const post = await postRepo.findById(req.params.id);  // handler querying DB directly
  if (post.author_id !== req.user.id) {                 // authorization is a business rule
    res.status(403).json({ error: "Forbidden" });
    return;
  }
  if (post.status !== "draft") {                        // business state rule in handler
    res.status(422).json({ error: "Only drafts can be published" });
    return;
  }
  await postRepo.update(post.id, { status: "published" }); // handler bypassing service layer
  res.json({ success: true });
}
```

**Correct:**

```typescript
// GOOD: handler delegates all logic to service
export async function publishPostHandler(req: Request, res: Response): Promise<void> {
  try {
    const result = await postService.publishPost(req.params.id, req.user.id);
    res.json({ id: result.id, status: result.status });
  } catch (e) {
    if (e instanceof NotFoundError) { res.status(404).json({ error: e.message }); return; }
    if (e instanceof ForbiddenError) { res.status(403).json({ error: e.message }); return; }
    if (e instanceof InvalidStateError) { res.status(422).json({ error: e.message }); return; }
    res.status(500).json({ error: "Internal error" });
  }
}

// GOOD: service contains all business logic
async publishPost(postId: string, requestorId: string): Promise<PostRecord> {
  const post = await this.postRepo.findById(postId);
  if (!post) throw new NotFoundError(`Post ${postId} not found`);
  if (post.authorId !== requestorId) throw new ForbiddenError("Only the author can publish");
  if (post.status !== "draft") throw new InvalidStateError("Only drafts can be published");
  return this.postRepo.save({ ...post, status: "published", publishedAt: new Date() });
}
```

---

## Framework Examples

**Express.js (TypeScript) — full wiring:**

```typescript
// src/index.ts
const postRepo = new PostgresPostRepository(db);
const notifier = new SendGridNotificationService(sendgridClient);
const postService = new PostService(postRepo, userRepo, notifier);
const postRouter = buildPostRouter(postService);
app.use("/api/posts", postRouter);
```

**FastAPI (Python) — dependency injection:**

```python
# app/dependencies.py
def get_post_service(db: Session = Depends(get_db)) -> PostService:
    return PostService(PostgresPostRepository(db), SendGridNotificationService())

# app/routers/posts.py
@router.post("/posts")
async def create_post(payload: CreatePostRequest, service: PostService = Depends(get_post_service)):
    ...
```

**Go (net/http) — constructor injection:**

```go
// cmd/server/main.go
postRepo := repository.NewPostgresPostRepository(db)
notifier := notification.NewSMTPNotificationService(smtpConfig)
postService := service.NewPostService(postRepo, userRepo, notifier)
postHandler := handler.NewPostHandler(postService)
mux.HandleFunc("POST /posts", postHandler.Create)
```

The wiring pattern is identical across all three: construct repositories, inject into services, inject services into handlers.

---

## Testing Strategy

**Unit tests — Service layer:**

- Test each service method by injecting in-memory or mock repositories and external clients.
- No database, no HTTP server, no external calls.
- The service must be constructable with fakes that implement the repository interfaces.

**Integration tests — Repository layer:**

- Test each repository method against a real database (test container, SQLite in-memory, or a dedicated test schema).
- Confirm SQL queries return the right data for given inputs.
- These tests do not exercise service logic.

**Acceptance tests — API-level via Gherkin:**

- Feature files describe the expected API behaviour from the outside.
- Step definitions send HTTP requests (supertest for Node, `httptest` for Go, `pytest` + `httpx` for Python) and assert on responses.
- These tests run against a full application instance with a test database.

```
tests/
  unit/
    services/
      PostService.test.ts        # or post_service_test.py / post_service_test.go
      UserService.test.ts
  fakes/
    InMemoryPostRepository.ts    # implements PostRepository interface
    InMemoryUserRepository.ts
    FakeNotificationService.ts
  integration/
    repositories/
      PostgresPostRepository.test.ts
      PostgresUserRepository.test.ts
  acceptance/
    create-post.steps.ts
    publish-post.steps.ts

specs/
  create-post.feature
  publish-post.feature
```
