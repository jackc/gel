require 'rspec'
require 'fileutils'

$tmpdir = 'spec/tmp/' + Time.now.strftime("%Y-%m-%d-%H-%M-%S")
FileUtils.mkdir_p $tmpdir

RSpec.configure do |config|
  def gst(args="")
    lib = File.expand_path(File.join(File.dirname(__FILE__), '..', 'lib'))
    bin = File.expand_path(File.join(File.dirname(__FILE__), '..', 'bin', 'gst'))
    `ruby -I #{lib} #{bin} #{args}`
  end
end

describe "gst" do
  it "emits static content unchanged" do
    gst("spec/examples/hello_world/hello_world.gst > #{$tmpdir}/hello_world.go")
    File.open("#{$tmpdir}/main.go", "w") do |f|
      f.puts <<-GO
package main

import (
  "bytes"
  "fmt"
)

func main() {
  var b bytes.Buffer
  HelloWorld(&b)
  fmt.Print(b.String())
}
      GO
    end
    output = `cd #{$tmpdir}; go run main.go hello_world.go`
    output.should eq "Hello, World!\n"
  end

  it "executes go code" do
    gst("spec/examples/hey_hey_hey/hey_hey_hey.gst > #{$tmpdir}/hey_hey_hey.go")
    File.open("#{$tmpdir}/main.go", "w") do |f|
      f.puts <<-GO
package main

import (
  "bytes"
  "fmt"
)

func main() {
  var b bytes.Buffer
  HeyHeyHey(&b)
  fmt.Print(b.String())
}
      GO
    end
    output = `cd #{$tmpdir}; go run main.go hey_hey_hey.go`
    output.should eq "Hey! Hey! Hey! \n"
  end

  it "interpolates strings" do
    gst("spec/examples/string_interpolation/string_interpolation.gst > #{$tmpdir}/string_interpolation.go")
    File.open("#{$tmpdir}/main.go", "w") do |f|
      f.puts <<-GO
package main

import (
  "bytes"
  "fmt"
)

func main() {
  var b bytes.Buffer
  StringInterpolation(&b)
  fmt.Print(b.String())
}
      GO
    end
    output = `cd #{$tmpdir}; go run main.go string_interpolation.go`
    output.should eq "Hello, Jack!\n"
  end
end
