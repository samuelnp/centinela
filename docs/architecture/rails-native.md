<!-- centinela:doc-version=1 template=docs/architecture/rails-native.md -->
# Rails-native Architecture Guide

> This document describes the **Rails-native archetype**. It is one of five supported architecture archetypes in Centinela. See [architecture-overview.md](architecture-overview.md) to confirm this is the right pattern for your project before reading further. Although the name says "Rails", this archetype applies equally to Django (Python) and Laravel (PHP) — the conventions map directly.

---

## When to Choose Rails-native

Choose Rails-native when:

- You are building a Rails, Django, or Laravel application
- Your domain maps naturally to database tables with CRUD operations
- The framework's conventions (Active Record, ORM, routing) already solve your architecture problems

Do NOT choose Rails-native when:

- Your domain logic cannot be expressed as database-backed models (use Hexagonal)
- You are building a stateless API service with no framework conventions (use N-Tier)
- You find yourself adding `Repository` classes, `Port` interfaces, and `UseCase` objects on top of Rails — you are fighting the framework

---

## Core Idea

**Active Record is not a violation — it is the design.**

The defining feature of Rails, Django, and Laravel is that the model is both the domain object and the database gateway. This is intentional. Martin Fowler's Active Record pattern explicitly places business logic alongside persistence. Layering Hexagonal ports and adapters on top of Rails is not "clean architecture" — it is overengineering that discards the framework's strongest advantages.

The Rails-native archetype works with this design:

- Models own business logic (validations, scopes, domain methods)
- Controllers are thin dispatchers — they parse requests and delegate
- Views display data — they do not compute or query
- Services (POROs/POPOs) handle operations too complex for a single model

**The risk to guard against is not Active Record — it is fat controllers and logic-heavy views.**

---

## Layer Details

### Views / Templates

The outermost layer. Renders HTML (or JSON) from data prepared by controllers. Contains no logic beyond simple iteration and conditional display.

**What it contains:**

- ERB templates (Rails), Jinja2 templates (Django), Blade templates (Laravel)
- Partials / template fragments for reusable layout pieces
- Helpers and view components for display-specific formatting

```
app/views/                         # Rails
  posts/
    index.html.erb
    show.html.erb
    new.html.erb
    _post_card.html.erb            # partial
  comments/
    _comment.html.erb
  shared/
    _navbar.html.erb
    _flash.html.erb
  layouts/
    application.html.erb

templates/                         # Django (example)
  posts/
    list.html
    detail.html
    _post_card.html
  base.html

resources/views/                   # Laravel (example)
  posts/
    index.blade.php
    show.blade.php
    partials/
      post-card.blade.php
```

**Rules:**

1. No database queries in templates. If a template calls `Post.where(...)`, that query belongs in the controller or model.
2. No conditionals that encode business rules. `{% if post.approved and post.author.is_premium %}` is a business rule — move it to a model method like `post.displayable_to_premium_users?`.
3. Helpers and view components format data for display. They do not compute business outcomes.
4. Templates receive objects from the controller — they do not call class methods to fetch data.
5. No I18n keys used only in one template defined inline — all user-facing strings go through the i18n system.
6. Partials accept local variables, not global state lookups.

---

### Controllers

Thin dispatchers. A controller action has one job: receive a request, call the appropriate model or service, and render the result.

**What it contains:**

- Request parsing (params, headers, authentication)
- Authorization checks (before_action, middleware)
- One call to a model or service
- Response rendering (redirect, render, JSON response)

```
app/controllers/                   # Rails
  application_controller.rb
  posts_controller.rb
  comments_controller.rb
  users_controller.rb
  api/
    v1/
      posts_controller.rb
      users_controller.rb

views/                             # Django (example)
  posts/
    views.py                       # PostListView, PostDetailView, PostCreateView

app/Http/Controllers/              # Laravel (example)
  PostController.php
  CommentController.php
  UserController.php
```

**Rules:**

1. Controller actions must not contain business logic. If an action has more than one `if` statement about domain state, extract that logic to a model method or service.
2. No direct SQL or ORM queries in controllers beyond what a model scope provides. `Post.published.by_author(current_user)` is fine — it delegates to the model. `Post.where("published = true AND author_id = ?", current_user.id)` in a controller is not.
3. No data transformation in controllers — format data in the model, a presenter, or a serializer.
4. Controller actions are 10–15 lines maximum. If longer, a service extraction is needed.
5. Authorization logic belongs in a dedicated layer (Pundit policies in Rails, permission classes in Django) — not inline in actions.
6. Before-actions / middleware handle cross-cutting concerns (auth, logging) — they do not implement business rules.

---

### Models (Active Record)

The business and persistence layer combined. This is where domain logic lives in Rails-native architecture.

**What it contains:**

- Database-backed fields and associations
- Validations — what makes a record valid
- Scopes — named, reusable query conditions
- Domain methods — business logic that belongs to this model
- Callbacks — lifecycle hooks (use sparingly and only for persistence-related side effects)

```
app/models/                        # Rails
  application_record.rb
  post.rb                          # validations, scopes, domain methods
  comment.rb
  user.rb
  tag.rb
  concerns/
    publishable.rb                 # shared behaviour: publish!, unpublish!, published scope
    auditable.rb                   # shared behaviour: track created_by, updated_by

models/                            # Django (example)
  post.py                          # Post model with custom managers and methods
  comment.py
  user.py

app/Models/                        # Laravel (example)
  Post.php
  Comment.php
  User.php
```

**Rules:**

1. Scopes are the correct place for reusable query conditions. A scope named `published` is better than `Post.where(status: :published)` repeated across controllers.
2. Domain methods belong on the model that owns the data. `post.publish!`, `user.deactivate!`, `order.cancel!` are model methods, not service methods.
3. Callbacks (`before_save`, `after_create`) are permitted only for persistence side effects. Using `after_create` to send a welcome email is a violation — use a service or job.
4. Associations must be declared explicitly with the correct relationship type — no polymorphic associations without deliberate justification.
5. Fat models are acceptable up to the point where a model is doing work that belongs to multiple models. When a model imports other models to orchestrate them, extract a service.
6. Validations are mandatory for every field that has a constraint. Rely on model validations, not only database constraints.

---

### Services / Plain Objects (POROs / POPOs)

Plain Ruby Objects (POROs), Plain Old Python Objects (POPOs), or equivalent — framework-free classes that handle operations too complex for a single model.

**When to extract a service:**

- An operation touches more than one model and coordinates them
- An operation involves an external API call (payment, email, SMS)
- An operation has significant conditional logic that makes the model too large
- The same operation is called from multiple controllers

```
app/services/                      # Rails
  posts/
    post_publisher.rb              # orchestrates publish! + notification + audit log
    post_archiver.rb
  users/
    user_registrar.rb              # creates User + sends welcome email + creates default settings
    account_deactivator.rb
  payments/
    charge_processor.rb            # wraps Stripe API call + creates Payment record
    refund_issuer.rb

services/                          # Django (example)
  post_service.py
  user_service.py
  payment_service.py

app/Services/                      # Laravel (example)
  PostPublisher.php
  UserRegistrar.php
  PaymentProcessor.php
```

**Rules:**

1. Services are initialized with their dependencies (models, external clients) — no global state lookups inside a service method.
2. Each service has one public method that represents its single responsibility. `PostPublisher#call`, `UserRegistrar#call`.
3. Services return a result object or raise a domain-specific error — not raw `true/false`.
4. Services do not render views or interact with HTTP request/response objects.
5. Services may call models and external clients. They do not call controllers.
6. Services are not a dumping ground. If a service is growing beyond 80 lines, split it.

---

### Jobs / Workers

Async operations deferred to a background queue. Any operation that should not block a web request belongs here.

**What it contains:**

- Background jobs that perform a single operation asynchronously
- Scheduled / recurring tasks (cron-style)
- Retry logic configuration

```
app/jobs/                          # Rails (Sidekiq / GoodJob)
  send_welcome_email_job.rb
  generate_monthly_report_job.rb
  sync_external_inventory_job.rb
  purge_expired_sessions_job.rb

tasks/                             # Django (Celery example)
  email_tasks.py
  report_tasks.py
  sync_tasks.py

app/Jobs/                          # Laravel (Horizon example)
  SendWelcomeEmail.php
  GenerateMonthlyReport.php
```

**Rules:**

1. Jobs delegate to services or models — they do not contain business logic themselves.
2. Jobs are idempotent: running the same job twice must not produce harmful duplicate side effects.
3. Jobs accept only primitive serializable arguments (IDs, strings) — not model instances, which may be stale by the time the job runs.
4. Retry behaviour is configured at the job level with an explicit maximum retry count.
5. Long-running jobs emit progress signals (logs, status updates) for observability.

---

## Dependency Direction

```
Views / Templates
       ↓
   Controllers
       ↓
  Models / Services
       ↓
   Jobs / Workers
       ↓
 External Services
```

Higher layers call lower layers. Lower layers never call back up. A model never imports a controller. A service never renders a view.

---

## Forbidden Patterns (G2)

| Pattern | Why it is forbidden |
|---|---|
| DB query in a view/template | Views must receive ready data from the controller — no lazy loading in templates |
| Business logic in a controller action | Controllers dispatch; domain logic belongs in models or services |
| A model that imports another controller | Models are domain objects — they have no concept of HTTP |
| Service that calls `render` or accesses `params` | Services are framework-free — they must not touch request/response objects |
| `after_create` callback that sends email | Callbacks are for persistence side effects only — use a job |
| Scopes defined inline in controllers | Reusable query logic belongs on the model |
| Job that contains significant business logic | Jobs delegate to services — they are not service replacements |

---

## What "No Business Logic in Outer Layer" Means (G7)

In Rails-native, the "outer layer" is **views/templates and controllers**.

**Violations:**

```ruby
# BAD: business rule in a view helper
def display_price(post)
  if post.author.premium? && post.created_at > 30.days.ago
    post.price * 0.9  # discount logic in the view layer
  else
    post.price
  end
end

# BAD: business logic in a controller action
def publish
  @post = Post.find(params[:id])
  if @post.author == current_user && @post.comments.count > 0 && !@post.flagged?
    @post.update!(status: :published, published_at: Time.current)
    NotificationMailer.published(@post).deliver_later
  end
  redirect_to @post
end
```

**Correct approach:**

```ruby
# GOOD: controller delegates entirely to the service
def publish
  @post = Post.find(params[:id])
  result = Posts::PostPublisher.new(@post, current_user).call
  result.success? ? redirect_to(@post) : render(:edit, status: :unprocessable_entity)
end

# GOOD: service owns the logic
class Posts::PostPublisher
  def initialize(post, requestor)
    @post = post
    @requestor = requestor
  end

  def call
    return failure("Not authorized") unless @post.author == @requestor
    return failure("Post has no comments") if @post.comments.none?
    return failure("Post is flagged") if @post.flagged?

    @post.publish!
    SendPublishedNotificationJob.perform_later(@post.id)
    success
  end
end

# GOOD: model owns the state transition
class Post < ApplicationRecord
  def publish!
    update!(status: :published, published_at: Time.current)
  end
end
```

---

## Testing Strategy

**Unit tests — Models and Services:**

- Test model validations, scopes, and domain methods with factory-generated records.
- Test services by calling them with model instances or doubles for external clients.
- No HTTP requests, no full application stack.

**Integration tests — Controller + Model:**

- Test full request/response cycles against a real test database.
- Confirm the controller dispatches correctly and the response contains the right data.
- These are not acceptance tests — they test the wiring, not the user-visible behaviour.

**Acceptance tests — Feature specs:**

- Gherkin feature files describe user-visible behaviour in domain language.
- Step definitions drive the full stack (Capybara for Rails, Playwright/Selenium for Django and Laravel).
- One feature file per user-facing capability.

```
spec/                              # Rails (RSpec)
  models/
    post_spec.rb
    user_spec.rb
  services/
    posts/post_publisher_spec.rb
    users/user_registrar_spec.rb
  controllers/
    posts_controller_spec.rb
  jobs/
    send_welcome_email_job_spec.rb

tests/                             # Django (pytest)
  test_models.py
  test_services.py
  test_views.py

tests/                             # PHPUnit (Laravel)
  Unit/
    PostTest.php
    PostPublisherTest.php
  Feature/
    PostControllerTest.php

tests/acceptance/                  # shared pattern
  publish-post.steps.ts
  register-user.steps.ts

specs/
  publish-post.feature
  register-user.feature
```
