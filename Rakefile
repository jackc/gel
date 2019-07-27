begin
  require 'bundler'
  Bundler.setup
rescue LoadError
  puts 'You must `gem install bundler` and `bundle install` to run rake tasks'
end

require 'rspec/core/rake_task'

file 'gel' => ['main.go'] do
  sh 'go build'
end

desc 'Build gel'
task build: 'gel'

RSpec::Core::RakeTask.new(:spec)
task spec: :build
task :default => :spec
