package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/download-pack", func(c *gin.Context) {
		nome := c.DefaultQuery("nome", "Nome")
		cargo := c.DefaultQuery("cargo", "Cargo")
		celular := c.DefaultQuery("celular", "Telefone")
		email := c.DefaultQuery("email", "email@psienergy.com.br")

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		imagens := []string{
			"assets/Logo2.png",
			"assets/ISO.png",
		}

		for _, path := range imagens {
			file, err := os.Open(path)
			if err != nil {
				fmt.Printf("Erro ao abrir imagem %s: %v\n", path, err)
				continue
			}
			zipFile, _ := zipWriter.Create(filepath.Base(path))
			io.Copy(zipFile, file)
			file.Close()
		}

		// HTML FINAL - BLINDADO PARA OUTLOOK
		// Medidas Exatas: 5,33cm (201px) x 1,55cm (59px)
		const docTemplateHTML = `
		<html xmlns:o='urn:schemas-microsoft-com:office:office' xmlns:w='urn:schemas-microsoft-com:office:word' xmlns='http://www.w3.org/TR/REC-html40'>
		<head>
			<meta charset='utf-8'>
			<title>Assinatura PSI</title>
			<style>
				/* Fallback para visualizadores web */
				body { font-family: Verdana, sans-serif; font-size: 8pt; color: #858585; }
				
				/* Media Query para Dark Mode */
				@media only screen and (prefers-color-scheme: dark) {
					.texto-comum, .link-texto { color: #E0E0E0 !important; }
					.nome-destaque { color: #FF8C42 !important; }
					.linha-vertical { border-left-color: #E0E0E0 !important; }
				}
			</style>
		</head>
		<body style="margin:0; padding:0;">
			
			<table width="800" border="0" cellspacing="0" cellpadding="0" style="width:800px; border-collapse: collapse;">
				<tr valign="middle">
					
					<td width="220" align="right" valign="middle" style="padding-right: 15px;">
						<img src="Logo2.png" width="201" height="59" alt="PSI Energy" style="display:block; border:0; width:201px; height:59px;">
					</td>

					<td width="220" valign="middle" style="padding-right: 10px;">
						<span class="nome-destaque" style="font-family: Verdana, sans-serif; font-size: 9pt; color: #F37021; font-weight: bold;">{{.Nome}}</span><br>
						
						<span class="texto-comum" style="font-family: Verdana, sans-serif; font-size: 8pt; color: #858585; font-style: italic;">{{.Cargo}}</span><br>
						
						<span style="font-size: 3pt; line-height: 3pt; display:block;">&nbsp;</span>

						<span class="texto-comum" style="font-family: Verdana, sans-serif; font-size: 8pt; color: #858585;">{{.Celular}}</span><br>
						
						<a href="mailto:{{.Email}}" style="text-decoration:none;">
							<span class="link-texto" style="font-family: Verdana, sans-serif; font-size: 8pt; color: #858585; text-decoration:none;">{{.Email}}</span>
						</a>
					</td>

					<td width="200" valign="middle" style="padding-right: 15px;">
						<a href="https://www.psienergy.com.br" style="text-decoration:none;">
							<span class="link-texto" style="font-family: Verdana, sans-serif; font-size: 8pt; color: #858585; text-decoration:none; font-style: italic;">www.psienergy.com.br</span>
						</a><br>
						
						<span class="texto-comum" style="font-family: Verdana, sans-serif; font-size: 8pt; color: #858585; font-style: italic;">(11) 4807-0708</span><br>
						
						<span class="texto-comum" style="font-family: Verdana, sans-serif; font-size: 8pt; color: #858585; font-style: italic;">
							Av. Luiz Pellizzari 420 Distrito<br>
							Industrial Jundiaí/SP<br>
							CEP: 13.213-073
						</span>
					</td>

					<td width="1" class="linha-vertical" style="border-left: 1px solid #858585; font-size: 1px; line-height: 1px;">&nbsp;</td>

					<td width="180" align="center" valign="middle" style="padding-left: 15px;">
						<img src="ISO.png" width="160" height="59" alt="ISO" style="display:block; border:0; width:160px; height:59px;">
					</td>

				</tr>
			</table>
		</body>
		</html>`

		tmpl, err := template.New("doc").Parse(docTemplateHTML)
		if err != nil {
			c.String(http.StatusInternalServerError, "Erro no template")
			return
		}

		var docBuffer bytes.Buffer
		docBuffer.WriteString("\xEF\xBB\xBF")
		
		err = tmpl.Execute(&docBuffer, gin.H{
			"Nome":    nome,
			"Cargo":   cargo,
			"Celular": celular,
			"Email":   email,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Erro ao processar template")
			return
		}

		safeNome := strings.ReplaceAll(nome, " ", "_")
		nomeArquivoDoc := fmt.Sprintf("Assinatura_%s.doc", safeNome)
		
		docInZip, _ := zipWriter.Create(nomeArquivoDoc)
		docInZip.Write(docBuffer.Bytes())

		zipWriter.Close()

		filenameZip := fmt.Sprintf("Pack_Assinatura_%s.zip", safeNome)
		c.Header("Content-Disposition", "attachment; filename="+filenameZip)
		c.Header("Content-Type", "application/zip")
		c.Data(http.StatusOK, "application/zip", buf.Bytes())
	})

	r.Run(":8080")
}