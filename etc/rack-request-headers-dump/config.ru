class Application
  def call(env)
    status  = 200
    headers = { "Content-Type" => "application/json" }
    body    = [ env.select { |k, v| k.start_with?('HTTP_') }.map { |(k, v)| [k, v].join(' ') }.join("\n") ]

    [status, headers, body]
  end
end

run Application.new
