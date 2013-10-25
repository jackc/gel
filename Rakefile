begin
  require 'bundler'
  Bundler.setup
rescue LoadError
  puts 'You must `gem install bundler` and `bundle install` to run rake tasks'
end

require 'rspec/core/rake_task'

file 'gst' => ['main.go'] do
  sh 'go build'
end

desc 'Build gst'
task build: 'gst'

RSpec::Core::RakeTask.new(:spec)
task :default => :spec
