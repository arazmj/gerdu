#include <iostream>
#include <sstream>
#include <curlpp/cURLpp.hpp>
#include <curlpp/Easy.hpp>
#include <curlpp/Options.hpp>
#include <boost/property_tree/ptree.hpp>

#include <sstream>

using namespace std;

// RAII cleanup


// Send request and get a result.
// Here I use a shortcut to get it in a string stream ...

int main() {

    try
    {
        curlpp::Cleanup cURLppStartStop;
        curlpp::Easy post;

        post.setOpt(curlpp::options::Url("http://localhost/cache/Hello"));
        post.setOpt(new curlpp::options::PostFields("World"));
        post.setOpt(curlpp::options::Port(8080));
        post.perform();

        curlpp::Easy get;
        get.setOpt(curlpp::options::Url("http://localhost/cache/Hello"));
        std::stringstream response;
        get.setOpt(new curlpp::options::WriteStream(&response));
        get.setOpt(curlpp::options::Port(8080));
        get.perform();

        cout << "Hello = " << response.str() << endl;
    }
    catch(exception& e)
    {
        cerr << e.what() << endl;
    }
}
