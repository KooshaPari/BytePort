package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		
		var user models.User; var project *models.Project;   
 
		project, err := getRequestDetails(w,r)
		if err != nil {
			fmt.Println("err getting dets: ", err)
			http.Error(w, "Error Reading Request", http.StatusInternalServerError)
			return
		}
		fmt.Println("Proj: ",project)
		user = project.User
		// decrypt portDets
		decryptedPortEndpoint, err := lib.DecryptSecret(user.Portfolio.RootEndpoint)
		if err != nil {
			fmt.Println("Error decrypting endpoint:", err)
			http.Error(w, "Error decrypting endpoint", http.StatusInternalServerError)
			return
		}
		decryptedPortKey, err := lib.DecryptSecret(user.Portfolio.APIKey)
		if err != nil {
			fmt.Println("Error decrypting key:", err)
			http.Error(w, "Error decrypting key", http.StatusInternalServerError)
			return
		}
		fmt.Println("Dets: " + decryptedPortEndpoint + " ||| " +decryptedPortKey )
		 templateStruct, err := getTemplate(decryptedPortEndpoint, decryptedPortKey)
		// pull Portfolio API Format
		if err != nil {
			fmt.Println("error getting template")
			http.Error(w,"error getting template", http.StatusInternalServerError)
		}
		fmt.Println("Struct : ", templateStruct)
		genRequest := generatePrompt(templateStruct, project)
		fmt.Println("Prompt: ", genRequest)
		decryptedOAI, err := lib.DecryptSecret(user.OpenAICreds.APIKey)
		if err != nil {
			fmt.Println("Error decrypting key:", err)
			http.Error(w, "Error decrypting key", http.StatusInternalServerError)
			return
		}
		filledObject, err := requestFilledTemplate(genRequest, decryptedOAI)
		// Generate Response (hand info above to ai langchain)
		if err != nil {
			fmt.Println("Bad Prompt Req")
			http.Error(w,"Bad Prompt Req", http.StatusInternalServerError)
			return
		}
		fmt.Println("Success , OBJ: ", filledObject)
		err = sendToPortfolio(filledObject, decryptedPortEndpoint, decryptedPortKey)
		// post to portfolio
		// return\
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}
func sendToPortfolio(object string, endpoint string, key string)(error){
	Reqbytes, err := json.Marshal(object)
	if err != nil {
		return err}
	body := bytes.NewBuffer(Reqbytes)
	uri := endpoint + "byteport"
	fmt.Println("uri: ", uri)
	req, err := http.NewRequest("POST",endpoint+"byteport",body)
	if err != nil {
		fmt.Println("Err buiding req:", err)
		return err
	}
	_ , err = spinhttp.Send(req)
	if err != nil {
		fmt.Println("Err Sending req:", err)
		return err
	}
	fmt.Println("Success")
	return nil

}
func generatePrompt(template string, project *models.Project) (string ){
	base := `Given a project template and project information, generate a filled portfolio project object.

Input format:
1. Template: A JSON object with empty fields representing the portfolio project structure
2. Project info: Object containing details about the application being deployed

Rules:
1. For complex array fields (screenshots, links, etc.), use a single item following these priority rules:
   - If example data exists in the sample filled template, copy the first item's structure
   - If no example exists, use appropriate dummy data
   
2. For generated text fields:
   - name: Use the project name
   - slug: Generate from name (lowercase, hyphenated)
   - shortDescription: One-line summary of project
   - description: Multi-paragraph description using markdown
   - color: Use main tech stack color or "primary"
   
3. For technical references:
   - logo: Use technology's logo if available, otherwise placeholder
   - skills: Include main technologies used
   - type: Map project type to standard categories (Website, Mobile App, CLI Tool, etc.)

4. For dates:
   - Use actual project dates when available
   - Default to current date for "from" if not specified

5. For metadata fields (company, location, etc.):
   - Use provided info if available
   - Leave empty if not relevant

Sample API call (Filled Example with Unfilled template in template{} (do not include template attribute in response)):
%s

Project Information:
Name: %s
Description: %s
Platform: %s
Type: %s
Readme: 
%s
User Information:
Name: %s
 
Expected response: A filled template object matching the required structure. String-Representable Object usable as response body without modification`
	prompt := fmt.Sprintf(base,template, project.Name,project.Description, project.Platform, project.Type, project.Readme, project.User.Name)
	return prompt
}
func getTemplate(endpoint string, key string) (string, error) {
    uri := endpoint + "byteport"
    req, err := http.NewRequest("GET", uri, nil)
    if err != nil {
        return "", fmt.Errorf("error building request: %v", err)
    }
    req.Header.Set("Authorization", "Bearer "+key)
    
    resp, err := spinhttp.Send(req)
    if err != nil {
        return "", fmt.Errorf("error sending request: %v", err)
    }
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("error reading response body: %v", err)
    }

    // Decode base64
    decodedBytes, err := base64.StdEncoding.DecodeString(string(body))
    if err != nil {
        return "", fmt.Errorf("error decoding base64: %v", err)
    }

    // At this point decodedBytes contains the JSON string
    // No need to marshal since it's already a string representation of JSON
    return string(decodedBytes), nil
}
func getRequestDetails(w http.ResponseWriter, r *http.Request)(*models.Project,   error){
	fmt.Println("Getting Template Dets")
	var project models.Project
	body, err := io.ReadAll(r.Body)
        if err != nil {
            fmt.Println("Error reading request body: ", err)
            http.Error(w, "Error reading request body", http.StatusInternalServerError)
            return nil,err
        }
        defer r.Body.Close()
	fmt.Println("Parsing JSON...")
        err = json.Unmarshal(body, &project)
        if err != nil {
            http.Error(w, "Error parsing JSON", http.StatusBadRequest)
            return nil,err
        }
		return &project, nil
}
func requestFilledTemplate(prompt string, key string) (string, error) {
    response, err := lib.RequestChatCompletion(prompt, key)
    if err != nil {
        return "", fmt.Errorf("error getting chat completion: %v", err)
    }
    return response, nil
}
func main() {}
