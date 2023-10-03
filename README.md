# infracloud-golang
This Application exposes following endpoints

GET("/:shortURL", server.handleRedirect) ---> for using the short urls generated by /shorten
POST("/shorten", server.handleShorten) ---> for generation of short urls
GET("/metrics", server.handleMetrics) ---> for displaying the top three domains
GET("/viewAll", server.handleViewAll) ----> for viewing all urls generated till now
DELETE("/deleteAll", server.handleDeleteAll) ----> for deleting test records


##To use this application

clone the code from github using git@github.com:pahari09/infracloud-golang.git

Open the code from infracloud-golang directory in VS code or any editor

From the terminal Execute the following command --->
go mod tidy ---> this will download the required gin and redis go package for executing the code
go run .\main.go

This will start the server for accepting the request
Now we can use the URL shortening service

1. POST request through curl command
From CMD terminal
P:\GoLand\urlShortner\infracloud-golang>curl -X POST -H "Content-Type: application/json" -d "{\"originalURL\":\"https://www.microsoft.com\"}" http://localhost:8080/shorten
{"short_url":"cdb4d88d"} 
From Powershell:
   Invoke-WebRequest -Uri http://localhost:8080/shorten -Method POST -Body '{"originalURL":"https://www.google.com"}' -ContentType "application/json"
   
3. GET request for using short url:-
Open any browser
In the address bar type:- http://localhost:8080/cdb4d88d
You need to replace cdb4d88d with your own short URL generated during POST call.
This will redirect you to the original website if it exists other wise page not found error will be displayed.

4. GET request for metrics (Top 3 domains)
   In the address bar type:- http://localhost:8080/metrics
This will display the top domain in the following way

   {
   "top_domains": [
   {
   "Score": 7,
   "Member": "www.samsung.com"
   },
   {
   "Score": 5,
   "Member": "www.shopify.com"
   },
   {
   "Score": 4,
   "Member": "www.tesla.com"
   }
   ]
   }
5.  GET request for /viewAll
    In the address bar type:- http://localhost:8080/viewAll
    
This will display all the addresses shortened till now in the following way.

    {
    "url:027b9d03": "https://www.instagram.com",
    "url:033d2b7b": "https://www.oracle.com",
    "url:09f0bcc9": "https://www.amazon.com",
    "url:1868061f": "https://www.wikipedia.org",
    "url:1feab940": "https://www.reddit.com",
    "url:30dd373d": "https://www.adidas.com",
    "url:3700594f": "https://www.ebay.com",
    "url:3f78535a": "https://www.netflix.com",
    "url:53dbb2f8": "https://www.samsung.com",
    "url:5c59bc01": "https://www.adobe.com",
    "url:6acd1853": "https://www.slack.com",
    "url:736d3707": "https://www.apple.com",
    "url:79adddb8": "https://www.microsoft.com",
    "url:9c0394f7": "https://www.salesforce.com",
    "url:KZlOegPF": "https://www.samsung.com/up",
    "url:a6fbc09c": "https://www.shopify.com",
    "url:a974b238": "https://www.zoom.us",
    "url:c53526a0": "https://www.linkedin.com",
    "url:ca6305e8": "https://www.nike.com",
    "url:cffd855a": "https://www.spotify.com",
    "url:fbc48530": "https://www.tesla.com",
    }
7. DELETE for deleting all the records (for testing purpose)

Execute the following command in the CMD terminal

   curl -X DELETE http://localhost:8080/deleteAll
