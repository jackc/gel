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
    FileUtils.cp("spec/examples/hello_world/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go hello_world.go`
    output.should eq "Hello, World!\n"
  end

  it "executes go code" do
    gst("spec/examples/hey_hey_hey/hey_hey_hey.gst > #{$tmpdir}/hey_hey_hey.go")
    FileUtils.cp("spec/examples/hey_hey_hey/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go hey_hey_hey.go`
    output.should eq "Hey! Hey! Hey! \n"
  end

  it "interpolates strings" do
    gst("spec/examples/string_interpolation/string_interpolation.gst > #{$tmpdir}/string_interpolation.go")
    FileUtils.cp("spec/examples/string_interpolation/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go string_interpolation.go`
    output.should eq "Hello, Jack!\n"
  end

  it "interpolates integers" do
    gst("spec/examples/integer_interpolation/integer_interpolation.gst > #{$tmpdir}/integer_interpolation.go")
    FileUtils.cp("spec/examples/integer_interpolation/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go integer_interpolation.go`
    output.should eq "1, 2, 3, 4, 5\n"
  end

  it "escapes for HTML by default when escape html set" do
    gst("spec/examples/escape_html/escape_html.gst > #{$tmpdir}/escape_html.go")
    FileUtils.cp("spec/examples/escape_html/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go escape_html.go`
    output.should eq "<p>Hello, &lt;Jack&gt;!</p>\n"
  end

  it "includes parameters in function" do
    gst("spec/examples/parameters/parameters.gst > #{$tmpdir}/parameters.go")
    FileUtils.cp("spec/examples/parameters/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go parameters.go`
    output.should eq "Hello, Jack!\nHello, Jack!\nHello, Jack!\n"
  end


  it "includes imports" do
    gst("spec/examples/imports/imports.gst > #{$tmpdir}/imports.go")
    FileUtils.cp("spec/examples/imports/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go imports.go`
    output.should eq "1"
  end
end
