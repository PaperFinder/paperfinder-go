# paperfinder-go

You know what it is

## Dependencies
 - python3
    - tika
 - java
 - go
 - will to live
 
 ## API GUIDE 
- /finder
  Queries a question and returns a result, if one is found.  
  **Usage:** /finder?q=<question>&s=<subject>  
  **Rensponse:** It returns a json file with the following parameters:
   - Query: It is the question submitted
   - Found: Whether a match was found. This can be either: True,Partial,False
     (If a paper is found *only*)
   - Paper: Full paper name
   - QPL: Question paper link
   - MSL: Mark scheme link
   Todo:
   - QN: Question Number
   - Context: Context of the question
- /subjects
  **Usage:** /subjects  
  **Response:** A list of subjects seperated by ','  
 
## FILE READER.go
  usage:
    `go run _filereader/main.go crawl <link>` It will crawl the specified link for past papers. The more general the more past papers.  
    **Example:** `go run _filereader/main.go crawl https://www.physicsandmathstutor.com/past-papers/a-level-physics/`  
    
## Building

`go build`

OR

`go run main.go`

## Info

Server runs on port 8080 by default, will add a config later
