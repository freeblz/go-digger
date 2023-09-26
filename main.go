package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"github.com/xlzd/gotp"
)

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Erro ao carregar arquivo .env")
	}

	return os.Getenv(key)
}

func main() {
	SECRET := goDotEnvVariable("SECRET_KEY")
	USER := goDotEnvVariable("USERNAME_JD")
	PASS := goDotEnvVariable("PASSWORD_JD")

    if len(SECRET) == 0 {
        log.Fatal("Chave SECRET_KEY para autenticação em .env está vazio")
    }

    if len(USER) == 0 {
        log.Fatal("Chave USERNAME_JD para autenticação em .env está vazio")
    }

    if len(PASS) == 0 {
        log.Fatal("Chave PASSWORD_JD para autenticação em .env está vazio")
    }



	totp := gotp.NewDefaultTOTP(SECRET)

	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Env("LANGUAGE=pt_BR"),
		chromedp.Flag("lang", "pt_BR"),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	done := make(chan string, 1)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *browser.EventDownloadProgress:
			completed := "(unknown)"
			if ev.TotalBytes != 0 {
				completed = fmt.Sprintf("%0.2f%%", ev.ReceivedBytes/ev.TotalBytes*100.0)
			}
			log.Printf("state: %s, completed: %s\n", ev.State.String(), completed)
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.State.String()
				close(done)
			}
		}
	})

	wd, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	wd += "/Hectares"
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Iniciando!")
	err = chromedp.Run(ctx,
		chromedp.Navigate("https://jddig.deere.com/#"),
	)

	if err != nil {
		log.Fatal(err.Error())
	}

	oktaLogin := chromedp.Tasks{
		chromedp.WaitVisible(`#okta-sign-in > div.auth-content > div > div > div > a`, chromedp.NodeVisible),
		chromedp.Sleep(2 * time.Second),
		chromedp.SetValue(`#okta-signin-username`, USER, chromedp.ByID),
		chromedp.SetValue(`#okta-signin-password`, PASS, chromedp.ByID),
		chromedp.WaitVisible(`#okta-signin-submit`, chromedp.ByID),
		chromedp.Click(`#okta-signin-submit`, chromedp.ByID)}

	var teveQueSelecionar bool
	oktaActuallySetToOkta := chromedp.Tasks{
		chromedp.WaitVisible("#okta-sign-in",
			chromedp.ByQuery),
		chromedp.Sleep(5 * time.Second),
		chromedp.Evaluate(`
	            (function(){
	                let listOfAuthOptions = document.querySelectorAll('#okta-dropdown-options > ul > li > a')
	                if(listOfAuthOptions.length > 0){
	                    for (let item of listOfAuthOptions){
	                        if(item.text.toUpperCase().includes("OKTA")){
	                            item.click()
                                return true
	                        }
	                    }
	                }
                    return false
	            })()
	        `, &teveQueSelecionar),
	}

	var oktaOtp chromedp.Tasks

	jdLogin := chromedp.Tasks{
		chromedp.WaitVisible(`#okta-sign-in > div.auth-content`, chromedp.NodeVisible),
		chromedp.SetValue(`#okta-signin-username`, USER, chromedp.ByID),
		chromedp.SetValue(`#okta-signin-password`, PASS, chromedp.ByID),
		chromedp.Sleep(1 * time.Second),
		chromedp.Click(`#okta-signin-submit`, chromedp.ByID),
	}

	var temp string
	abrirTabelaDig := chromedp.Tasks{
		chromedp.WaitVisible("#eaCheckbox", chromedp.ByQuery),
		chromedp.Sleep(5 * time.Second),
		chromedp.Click("#eaCheckbox", chromedp.ByQuery),
		chromedp.WaitVisible("#dijit_form_Button_10 > span.dijitReset.dijitInline.dijitIcon.dijitIconTable", chromedp.BySearch),
		chromedp.Evaluate(`
            (function(){
                document.querySelector("#dijit_form_Button_10 > span.dijitReset.dijitInline.dijitIcon.dijitIconTable").click()
                return "Tabela do dig aberta"
            })()
        `, &temp),
	}

	selecionaOrganizacoes := chromedp.Tasks{
		chromedp.WaitVisible("#dijit_layout_TabContainer_0_tablist > div.dijitTabListWrapper.dijitTabContainerTopNone.dijitAlignCenter > div > div:nth-child(2)",
			chromedp.BySearch),
		chromedp.Sleep(6 * time.Second),
		chromedp.Click("#dijit_layout_TabContainer_0_tablist > div.dijitTabListWrapper.dijitTabContainerTopNone.dijitAlignCenter > div > div:nth-child(2)",
			chromedp.BySearch),
	}

	abreOpcoesAbreMenu := chromedp.Tasks{
		chromedp.Sleep(3 * time.Second),
		chromedp.Evaluate(`
            const clickEvent = new Event('click', {
            bubbles: true, 
              cancelable: true, 
            });
            const widget = dijit.byId("dijit_form_DropDownButton_1");
            widget._onDropDownMouseDown(clickEvent)`, nil),
	}

	iniciaDownload := chromedp.Tasks{
		chromedp.Sleep(2 * time.Second),
		chromedp.Evaluate(`
	    const items = document.querySelector("#dijit_DropDownMenu_2 > tbody").children
        items[items.length -1].click()
	    `, nil),
		chromedp.Sleep(3 * time.Second),
		browser.
			SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath(wd).
			WithEventsEnabled(true),
		chromedp.Click("#jimu_dijit_Message_0 > div.button-container > div:nth-child(3)", chromedp.ByQuery),
	}

	err = chromedp.Run(ctx, oktaLogin)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = chromedp.Run(ctx, oktaActuallySetToOkta)

	fmt.Printf("Precisou selecionar: %t\n", teveQueSelecionar)

	if teveQueSelecionar {
		oktaOtp = chromedp.Tasks{
			chromedp.WaitVisible(`#input73`, chromedp.ByID),
			chromedp.SetValue(`#input73`, totp.Now(), chromedp.ByID),
			chromedp.Click(`#form67 > div.o-form-button-bar > input`, chromedp.NodeVisible)}
	} else {
		oktaOtp = chromedp.Tasks{
			chromedp.WaitVisible(`#input72`, chromedp.ByID),
			chromedp.SetValue(`#input72`, totp.Now(), chromedp.ByID),
			chromedp.Click(`#form66 > div.o-form-button-bar > input`, chromedp.NodeVisible)}

	}

	err = chromedp.Run(ctx, oktaOtp)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = chromedp.Run(ctx, jdLogin)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = chromedp.Run(ctx, abrirTabelaDig)
	fmt.Println(temp)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = chromedp.Run(ctx, selecionaOrganizacoes)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = chromedp.Run(ctx, abreOpcoesAbreMenu)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = chromedp.Run(ctx, iniciaDownload)

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("wrote %s", filepath.Join(wd, <-done))

	// time.Sleep(3600 * time.Second)

}
