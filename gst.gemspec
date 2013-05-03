# coding: utf-8
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'gst/version'

Gem::Specification.new do |spec|
  spec.name          = "gst"
  spec.version       = Gst::VERSION
  spec.authors       = ["Jack Christensen"]
  spec.email         = ["jack@jackchristensen.com"]
  spec.description   = %q{Go Static Templates}
  spec.summary       = %q{GST is a templating library that compiles templates into Go functions.}
  spec.homepage      = "https://github.com/JackC/gst"
  spec.license       = "MIT"

  spec.files         = `git ls-files`.split($/)
  spec.executables   = spec.files.grep(%r{^bin/}) { |f| File.basename(f) }
  spec.test_files    = spec.files.grep(%r{^(test|spec|features)/})
  spec.require_paths = ["lib"]

  spec.add_development_dependency "bundler", "~> 1.3"
  spec.add_development_dependency "rake"
  spec.add_development_dependency 'rspec', '~> 2.13'
end
