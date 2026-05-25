Feature: Centinela landing page
  As a developer evaluating Centinela
  I want a fast, visual landing page at https://samuelnp.github.io/centinela/
  So I can understand the core idea in seconds and decide whether to install

  Scenario: Full interactive render shows all required above-the-fold elements
    Given a browser with JavaScript enabled opens https://samuelnp.github.io/centinela/
    When the page finishes loading
    Then the hero contains an img element whose src includes "logo-banner.png"
    And the hero contains the text "Centinela"
    And the hero contains the value prop "plan → code → tests → validate → docs — enforced"
    And the hero contains a code block with the exact text "go install github.com/samuelnp/centinela@latest"
    And the hero contains a link with text "Get Started"
    And the hero contains a link with text "Star on GitHub"

  Scenario: Install command exactly matches the README canonical string
    Given the rendered landing page HTML
    When a test inspects the install command block
    Then it contains the exact string "go install github.com/samuelnp/centinela@latest"
    And no other version tag or path variant appears in that block

  Scenario: Pipeline diagram section is present with five step labels
    Given the rendered landing page HTML
    When a test inspects the pipeline section
    Then it contains the text "plan"
    And it contains the text "code"
    And it contains the text "tests"
    And it contains the text "validate"
    And it contains the text "docs"

  Scenario: Greenfield roadmap section is present
    Given the rendered landing page HTML
    When a test inspects the page body
    Then a section depicting describe-project to roadmap to feature advancement is present
    And that section contains the concept of phases and features

  Scenario: Enforcement "aha" panel is present
    Given the rendered landing page HTML
    When a test inspects the page body
    Then a block depicting the prewrite hook blocking a write is present
    And that block contains a reason or "blocked" message

  Scenario: Absolute OG and Twitter meta tags are set
    Given the rendered landing page HTML
    When a test reads the head meta tags
    Then the og:image content attribute value starts with "https://samuelnp.github.io/centinela/assets/social-preview.png"
    And the twitter:image content attribute value starts with "https://samuelnp.github.io/centinela/assets/social-preview.png"
    And the og:url content attribute value is "https://samuelnp.github.io/centinela/"
    And the twitter:card content attribute value is "summary_large_image"

  Scenario: No external runtime CDN or JS framework dependency
    Given the rendered landing page HTML
    When a test scans all script and link elements
    Then no script element has a src attribute pointing to an external host
    And no link element with rel stylesheet has an href pointing to an external host

  Scenario: demo.gif is lazy-loaded with explicit dimensions and a placeholder
    Given the rendered landing page HTML
    When a test inspects the img element for assets/demo.gif
    Then it has a loading attribute equal to "lazy"
    And it has a decoding attribute equal to "async"
    And it has a non-empty width attribute
    And it has a non-empty height attribute

  Scenario: Footer contains real non-empty outbound links
    Given the rendered landing page HTML
    When a test inspects all anchor elements inside the footer
    Then every anchor has an href attribute that is not empty and does not equal "#"
    And at least one href contains "github.com/samuelnp/centinela"

  Scenario: No-JS degraded path — all content remains legible
    Given a browser with JavaScript disabled opens https://samuelnp.github.io/centinela/
    When the page finishes loading
    Then the hero text, value prop, and install command are visible
    And the pipeline section labels are visible
    And the greenfield section is visible
    And the enforcement panel is visible
    And the footer links are visible

  Scenario: Reduced-motion preference disables CSS animations
    Given a browser where prefers-reduced-motion is set to reduce
    When the landing page is loaded
    Then no CSS animation or transition plays on any element
    And all static content — hero, pipeline, greenfield, enforcement, footer — is still present

  Scenario: Narrow viewport reflows pipeline and roadmap without horizontal overflow
    Given a browser viewport width of 360 pixels
    When the landing page is loaded
    Then the pipeline diagram stacks vertically
    And the page has no horizontal scrollbar
    And no content is clipped or obscured

  Scenario: Missing demo.gif does not break page layout
    Given the rendered landing page
    When the demo.gif asset returns a 404 error
    Then the img element for demo.gif shows its alt text
    And the surrounding layout does not shift or collapse
    And no JavaScript error is thrown
