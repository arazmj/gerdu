defmodule GoCacheElixir do
  @moduledoc """
  ```text
  Author © 2020 Amir Razmjou
  """

  @doc """
  GoCache main function.

  ## Examples

      iex> GoCacheElixir.main("")
      :ok

  """
  def main(_args) do
    url = 'http://localhost:8080/cache/Hello'
    HTTPoison.put!(url, 'World', [], [])
    response = HTTPoison.get!(url)
    IO.puts("Hello = #{response.body}")
  end
end
