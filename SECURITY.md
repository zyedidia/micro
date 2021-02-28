# Security Policy

## Supported Versions

Use this section to tell people about which versions of your project are
currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 5.1.x   | :white_check_mark: |
| 5.0.x   | :x:                |
| 4.0.x   | :white_check_mark: |
| < 4.0   | :x:                |

## Reporting a Vulnerability

Use this section to tell people how to report a vulnerability.

Tell them where to go, how often they can expect to get an update on a
reported vulnerab$: << "lib" and require "scientist/version"

Gem::Specification.new do |gem|
  gem.name          = "scientist"
  gem.description   = "A Ruby library for carefully refactoring critical paths"
  gem.version       = Scientist::VERSION
  gem.authors       = ["GitHub Open Source", "John Barnette", "Rick Bradley", "Jesse Toth", "Nathan Witmer"]
  gem.email         = ["opensource+scientist@github.com", "jbarnette@github.com", "rick@rickbradley.com", "jesseplusplus@github.com","zerowidth@github.com"]
  gem.summary       = "Carefully test, measure, and track refactored code."
  gem.homepage      = "https://github.com/github/scientist"
  gem.license       = "MIT"

  gem.required_ruby_version = '>= 2.3'

  gem.files         = `git ls-files`.split($/)
  gem.executables   = []
  gem.test_files    = gem.files.grep(/^test/)
  gem.require_paths = ["lib"]

  gem.add_development_dependency "minitest", "~> 5.8"
  gem.add_development_dependency "coveralls", "~> 0.8"
end
ility, what to expect if the vulnerability is accepted or
declined, etc.
