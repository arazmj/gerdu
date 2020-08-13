using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Threading.Tasks;
using static System.Net.WebRequestMethods.Http;

namespace CSharp
{
    class Program
    {
        private static string _hostname = "http://localhost";
        private static string _port = "8080";
        static void Main(string[] args)
        {
            var client = new HttpClient();

            client.PostAsync($"{_hostname}:{_port}/cache/Hello",
                new StringContent("World")).Wait();

            var response = client.GetAsync($"{_hostname}:{_port}/cache/Hello")
                .Result;
            
            if (response != null)
            {
                var value = string.Empty;

                Task task = response.Content.ReadAsStreamAsync().ContinueWith(t =>
                {
                    var stream = t.Result;
                    using (var reader = new StreamReader(stream))
                    {
                        value = reader.ReadToEnd();
                    }
                });

                task.Wait();
                Console.WriteLine($"Hello = {value}");
            }

        }
    }
}