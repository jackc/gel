require "yaml"
require "erb"

module Gst
  class Template
    attr_reader :source
    attr_reader :metadata
    attr_reader :body

    def initialize(source)
      @source = source
      parse
    end

    def render
      erb_template.result(binding)
    end

    private
    def parse
      header, @body = @source.split("---\n")
      parse_header(header)
    end

    def parse_header(header)
      @metadata = YAML.load(header)
    end

    def package
      metadata.fetch("package")
    end

    def imports
      additional_imports = metadata.fetch("imports", "").split.map(&:strip)
      (["io"] + segments.map(&:imports).flatten.compact + additional_imports).uniq
    end

    def func
      metadata.fetch("func")
    end

    def parameters
      default = "writer io.Writer"
      if additional_parameters = metadata["parameters"]
        "#{default}, #{additional_parameters}"
      else
        default
      end
    end

    def escape
      metadata["escape"]
    end

    def new_default_string_segment(content)
      if escape == "html"
        HTMLEscapedStringInterpolationSegment.new content
      elsif escape == nil
        StringInterpolationSegment.new content
      else
        raise "Unknown escape type"
      end
    end

    def erb_template
      return @erb_template if defined?(@erb_template)
      path = File.expand_path(File.join(File.dirname(__FILE__), 'template.erb'))
      data = File.read path
      @erb_template = ERB.new(data, nil, '-')
    end

    def segments
      return @segments if defined?(@segments)

      @segements = body.scan(/<%.*?%>|(?:[^<]|<(?!%))+/).map do |s|
        if m = s[/(?<=\A<%=i).*(?=%>\Z)/]
          IntegerInterpolationSegment.new m
        elsif m = s[/(?<=\A<%=).*(?=%>\Z)/]
          new_default_string_segment(m)
        elsif m = s[/(?<=\A<%).*(?=%>\Z)/]
          GoSegment.new m
        else
          StringSegment.new s
        end
      end
    end
  end

  class Segment
    def initialize(content)
      @content = content
    end

    def imports
      []
    end
  end

  class StringSegment < Segment
    def to_go
      "io.WriteString(writer, `#{@content}`)"
    end
  end

  class GoSegment < Segment
    def to_go
      @content
    end
  end

  class StringInterpolationSegment < Segment
    def to_go
      "io.WriteString(writer, #{@content})"
    end
  end

  class HTMLEscapedStringInterpolationSegment < Segment
    def to_go
      "io.WriteString(writer, html.EscapeString(#{@content}))"
    end

    def imports
      %w[html]
    end
  end

  class IntegerInterpolationSegment < Segment
    def to_go
      "io.WriteString(writer, strconv.FormatInt(int64(#{@content}), 10))"
    end

    def imports
      %w[strconv]
    end
  end
end
