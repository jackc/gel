begin
  require 'bundler'
  Bundler.setup
rescue LoadError
  puts 'You must `gem install bundler` and `bundle install` to run rake tasks'
end

require "rake/testtask"

file 'gel' => ['main.go'] do
  sh 'go build'
end

desc 'Build gel'
task build: 'gel'

task test: :build
Rake::TestTask.new(:test) do |t|
  t.libs << "test"
  t.test_files = FileList['test/**/*_test.rb']
end

task default: :test
