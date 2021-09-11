package worker

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

type WorkerError struct {
	Err          error
	FunctionName string
}

func (we *WorkerError) Error() string {
	return fmt.Sprintf("%s : %s", we.FunctionName, we.Err.Error())
}

func Auth(wd selenium.WebDriver, login, password string) error {
	werr := &WorkerError{}
	werr.FunctionName = "Auth"

	werr.Err = wd.Get("https://remanga.org/")
	if werr.Err != nil {
		return werr
	}
	time.Sleep(1 * time.Second)

	var elem selenium.WebElement
	elem, werr.Err = wd.FindElement(selenium.ByCSSSelector, "button.MuiButtonBase-root.c30")
	if werr.Err != nil {
		return werr
	}
	werr.Err = elem.Click()
	if werr.Err != nil {
		return werr
	}
	time.Sleep(1 * time.Second)

	elem, werr.Err = wd.FindElement(selenium.ByCSSSelector, "#login")
	if werr.Err != nil {
		return werr
	}
	werr.Err = elem.SendKeys(login)
	if werr.Err != nil {
		return werr
	}

	elem, werr.Err = wd.FindElement(selenium.ByCSSSelector, "#password")
	if werr.Err != nil {
		return werr
	}
	werr.Err = elem.SendKeys(password)
	if werr.Err != nil {
		return werr
	}

	elem, werr.Err = wd.FindElement(selenium.ByCSSSelector, `button.MuiButtonBase-root.MuiButton-root[type="submit"]`)
	if werr.Err != nil {
		return werr
	}
	werr.Err = elem.Click()
	if werr.Err != nil {
		return werr
	}

	return nil
}

func Prepare(wd selenium.WebDriver, pathToImage string) error {
	werr := &WorkerError{}
	werr.FunctionName = "Prepare"

	werr.Err = wd.Get("https://remanga.org/panel/add-titles/")
	if werr.Err != nil {
		return werr
	}
	time.Sleep(1 * time.Second)

	var elem selenium.WebElement
	elem, werr.Err = wd.FindElement(selenium.ByCSSSelector, "#id_cover")
	if werr.Err != nil {
		return werr
	}

	werr.Err = elem.SendKeys(pathToImage)
	if werr.Err != nil {
		return werr
	}

	return nil
}

func Post(wd selenium.WebDriver, titleType, ruName, enName, krName, link string) error {
	werr := &WorkerError{}
	werr.FunctionName = "Post"

	js := fmt.Sprintf(`
		var file = $("#id_cover")[0].files[0];
		var formData = new FormData();
		formData.append("cover", file);
		formData.append("csrfmiddlewaretoken", document.getElementsByName("csrfmiddlewaretoken")[0].value);
		formData.append("en_name", "%s");
		formData.append("rus_name", "%s");
		formData.append("another_name", "%s");
		formData.append("description", "");
		formData.append("type", "%s");
		formData.append("categories", "5");
		formData.append("categories", "6");
		formData.append("genres", "2");
		formData.append("genres", "38");
		formData.append("publishers", "2560");
		formData.append("status", "4");
		formData.append("age_limit", "0");
		formData.append("issue_year", "2021");
		formData.append("mangachan_link", "");
		formData.append("original_link", "%s");
		formData.append("anlate_link", "");
		formData.append("readmanga_link", "");
		formData.append("user_message", "");
		var xhr = new XMLHttpRequest();
		xhr.open('POST', '/panel/add-titles/', true);
		xhr.send(formData);
	`, enName, ruName, krName, titleType, link)

	_, werr.Err = wd.ExecuteScript(js, nil)
	if werr.Err != nil {
		return werr
	}
	return nil
}
