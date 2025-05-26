package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"

	"kursovaya/stego"
)

const (
	AlgorithmFractal = "Фрактал"
)

type StegoApp struct {
	app                      fyne.App
	window                   fyne.Window
	coverImagePath           *widget.Entry
	secretDataPath           *widget.Entry
	stegoImagePath           *widget.Entry
	outputPath               *widget.Entry
	algorithm                *widget.RadioGroup
	embeddingRate            *widget.Slider
	fractalType              *widget.Select
	fractalIterations        *widget.Entry
	fractalThreshold         *widget.Entry
	coverImagePreview        *canvas.Image
	stegoImagePreview        *canvas.Image
	metricsText              *widget.Label
	originalImagePath        *widget.Entry
	stegoImagePathForMetrics *widget.Entry
	fractalParamsGroup       *fyne.Container
}

func NewStegoApp() *StegoApp {
	a := app.New()
	w := a.NewWindow("Стеганография с фракталами")
	w.Resize(fyne.NewSize(1200, 700))

	stegApp := &StegoApp{
		app:    a,
		window: w,
	}

	stegApp.createUI()
	return stegApp
}

func (a *StegoApp) createUI() {
	tabs := container.NewAppTabs(
		a.createEmbedTab(),
		a.createExtractTab(),
		a.createMetricsTab(),
	)

	a.window.SetContent(tabs)
}

func (a *StegoApp) createEmbedTab() *container.TabItem {
	// Cover Image Selection
	a.coverImagePath = widget.NewEntry()
	coverBrowse := widget.NewButton("Выбрать", a.browseCoverImage)
	coverBrowse.Resize(fyne.NewSize(120, 38))

	// Secret Data Selection
	a.secretDataPath = widget.NewEntry()
	secretBrowse := widget.NewButton("Выбрать", a.browseSecretData)
	secretBrowse.Resize(fyne.NewSize(120, 38))

	// Algorithm Selection
	a.algorithm = widget.NewRadioGroup([]string{AlgorithmFractal}, nil)
	a.algorithm.SetSelected(AlgorithmFractal)
	a.algorithm.OnChanged = a.toggleFractalParams

	// Fractal Parameters
	a.fractalType = widget.NewSelect([]string{"Мандельброт", "Жулиа"}, nil)
	a.fractalType.SetSelected("Мандельброт")
	a.fractalIterations = widget.NewEntry()
	a.fractalIterations.SetText("100")
	a.fractalThreshold = widget.NewEntry()
	a.fractalThreshold.SetText("2.0")

	a.fractalParamsGroup = container.NewVBox(
		widget.NewLabel("Параметры фрактала:"),
		container.NewVBox(
			widget.NewLabel("Тип фрактала:"),
			a.fractalType,
			widget.NewLabel("Итерация:"),
			a.fractalIterations,
			widget.NewLabel("Порог:"),
			a.fractalThreshold,
		),
	)
	a.fractalParamsGroup.Hide()

	// Embedding Rate
	a.embeddingRate = widget.NewSlider(0.1, 0.9)
	a.embeddingRate.Value = 0.4
	rateLabel := widget.NewLabel(fmt.Sprintf("Коэффициент встраивания: %.1f", a.embeddingRate.Value))
	a.embeddingRate.OnChanged = func(v float64) {
		rateLabel.SetText(fmt.Sprintf("Коэффициент встраивания: %.1f", v))
	}

	// Output Path
	a.outputPath = widget.NewEntry()
	outputBrowse := widget.NewButton("Выбрать", a.browseOutput)
	outputBrowse.Resize(fyne.NewSize(120, 38))

	// Embed Button
	embedButton := widget.NewButton("Встроить данные", a.embedData)

	// Image Previews
	a.coverImagePreview = canvas.NewImageFromResource(nil)
	a.coverImagePreview.SetMinSize(fyne.NewSize(200, 200))
	a.stegoImagePreview = canvas.NewImageFromResource(nil)
	a.stegoImagePreview.SetMinSize(fyne.NewSize(200, 200))

	previews := container.NewHBox(
		container.NewVBox(widget.NewLabel("Стеганографический контейнер:"), a.coverImagePreview),
		container.NewVBox(widget.NewLabel("Изображение для встраивания:"), a.stegoImagePreview),
	)

	// Form Layout
	form := container.NewVBox(
		widget.NewLabel("Стеганографический контейнер:"),
		container.New(&weightedLayout{leftWeight: 0.8, rightWeight: 0.2}, a.coverImagePath, coverBrowse),
		widget.NewLabel("Файл для встраивания:"),
		container.New(&weightedLayout{leftWeight: 0.8, rightWeight: 0.2}, a.secretDataPath, secretBrowse),
		widget.NewLabel("Алгоритм встраивания:"),
		a.algorithm,
		a.fractalParamsGroup,
		rateLabel,
		a.embeddingRate,
		widget.NewLabel("Выходной файл:"),
		container.New(&weightedLayout{leftWeight: 0.8, rightWeight: 0.2}, a.outputPath, outputBrowse),
		embedButton,
		previews,
	)

	return container.NewTabItem("Встроить данные", form)
}

func (a *StegoApp) createExtractTab() *container.TabItem {
	// Stego Image Selection
	a.stegoImagePath = widget.NewEntry()
	stegoBrowse := widget.NewButton("Выбрать", a.browseStegoImage)
	stegoBrowse.Resize(fyne.NewSize(120, 38))

	// Output Path
	outputPath := widget.NewEntry()
	outputBrowse := widget.NewButton("Выбрать", a.browseOutput)
	outputBrowse.Resize(fyne.NewSize(120, 38))

	// Algorithm Selection
	algorithm := widget.NewRadioGroup([]string{AlgorithmFractal}, nil)
	algorithm.SetSelected(AlgorithmFractal)

	// Extract Button
	extractButton := widget.NewButton("Извлечь данные", a.extractData)

	// Form Layout
	form := container.NewVBox(
		widget.NewLabel("Входное изображение:"),
		container.New(&weightedLayout{leftWeight: 0.8, rightWeight: 0.2}, a.stegoImagePath, stegoBrowse),
		widget.NewLabel("Выходной файл с данными:"),
		container.New(&weightedLayout{leftWeight: 0.8, rightWeight: 0.2}, outputPath, outputBrowse),
		widget.NewLabel("Алгоритм извлечения:"),
		algorithm,
		extractButton,
	)

	return container.NewTabItem("Извлечь данные", form)
}

func (a *StegoApp) createMetricsTab() *container.TabItem {
	// Original and Stego Image Selection
	a.originalImagePath = widget.NewEntry()
	originalBrowse := widget.NewButton("Выбрать", func() {
		a.browseImage(a.originalImagePath)
	})
	originalBrowse.Resize(fyne.NewSize(120, 38))

	a.stegoImagePathForMetrics = widget.NewEntry()
	stegoBrowse := widget.NewButton("Выбрать", func() {
		a.browseImage(a.stegoImagePathForMetrics)
	})
	stegoBrowse.Resize(fyne.NewSize(120, 38))

	// Calculate Metrics Button
	metricsButton := widget.NewButton("Посчитать метрики встраивания", a.calculateMetrics)

	// Metrics Display
	a.metricsText = widget.NewLabel("")
	a.metricsText.Wrapping = fyne.TextWrapWord

	// Visualization would require more complex implementation in Fyne
	// For simplicity, we'll just show the metrics

	// Form Layout
	form := container.NewVBox(
		widget.NewLabel("Оригинальное изображение:"),
		container.New(&weightedLayout{leftWeight: 0.8, rightWeight: 0.2}, a.originalImagePath, originalBrowse),
		widget.NewLabel("Стеганографическое изображение:"),
		container.New(&weightedLayout{leftWeight: 0.8, rightWeight: 0.2}, a.stegoImagePathForMetrics, stegoBrowse),
		metricsButton,
		widget.NewLabel("Результаты метрик:"),
		a.metricsText,
	)

	return container.NewTabItem("Анализ метрик", form)
}

func (a *StegoApp) toggleFractalParams(s string) {
	if s == AlgorithmFractal {
		a.fractalParamsGroup.Show()
	} else {
		a.fractalParamsGroup.Hide()
	}
}

func (a *StegoApp) browseCoverImage() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err == nil && reader != nil {
			a.coverImagePath.SetText(reader.URI().Path())
			a.loadImagePreview(reader.URI().Path(), a.coverImagePreview)
		}
	}, a.window)
}

func (a *StegoApp) browseSecretData() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err == nil && reader != nil {
			a.secretDataPath.SetText(reader.URI().Path())
		}
	}, a.window)
}

func (a *StegoApp) browseStegoImage() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err == nil && reader != nil {
			a.stegoImagePath.SetText(reader.URI().Path())
		}
	}, a.window)
}

func (a *StegoApp) browseOutput() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err == nil && writer != nil {
			a.outputPath.SetText(writer.URI().Path())
		}
	}, a.window)
}

func (a *StegoApp) browseImage(entry *widget.Entry) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err == nil && reader != nil {
			entry.SetText(reader.URI().Path())
		}
	}, a.window)
}

func (a *StegoApp) loadImagePreview(path string, img *canvas.Image) {
	file, err := os.Open(path)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	i, _, err := image.Decode(file)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Resize for preview
	resized := imaging.Resize(i, 200, 200, imaging.Lanczos)
	img.Image = resized
	img.Refresh()
}

func (a *StegoApp) embedData() {
	// Check if all required fields are filled
	if a.coverImagePath.Text == "" {
		dialog.ShowError(errors.New("please select a cover image"), a.window)
		return
	}
	if a.secretDataPath.Text == "" {
		dialog.ShowError(errors.New("please select a secret data file"), a.window)
		return
	}
	if a.outputPath.Text == "" {
		dialog.ShowError(errors.New("please specify an output path"), a.window)
		return
	}

	// Load the cover image
	coverImage, err := loadImage(a.coverImagePath.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load cover image: %w", err), a.window)
		return
	}

	// Load the secret data
	secretData, err := os.ReadFile(a.secretDataPath.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load secret data: %w", err), a.window)
		return
	}

	// Create steganography config
	config := stego.Config{
		EmbeddingRate: a.embeddingRate.Value,
	}

	// Set fractal parameters if needed
	if a.algorithm.Selected == AlgorithmFractal {
		// Parse fractal parameters
		iterations, err := strconv.Atoi(a.fractalIterations.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid iterations value: %w", err), a.window)
			return
		}

		threshold, err := strconv.ParseFloat(a.fractalThreshold.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid threshold value: %w", err), a.window)
			return
		}

		config.FractalParams = &stego.FractalParams{
			Type:       a.fractalType.Selected,
			Iterations: iterations,
			Threshold:  threshold,
		}
	}

	// Get the appropriate steganography algorithm
	algorithm, err := stego.Factory(a.algorithm.Selected)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Embed the data
	stegoImage, err := algorithm.Embed(coverImage, secretData, config)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to embed data: %w", err), a.window)
		return
	}

	// Save the stego image
	err = saveImage(a.outputPath.Text, stegoImage)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save stego image: %w", err), a.window)
		return
	}

	// Update the preview
	a.loadImagePreview(a.outputPath.Text, a.stegoImagePreview)

	dialog.ShowInformation("Успех", "Данные успешно сокрыты", a.window)
}

func (a *StegoApp) extractData() {
	// Check if all required fields are filled
	if a.stegoImagePath.Text == "" {
		dialog.ShowError(errors.New("please select a stego image"), a.window)
		return
	}
	if a.outputPath.Text == "" {
		dialog.ShowError(errors.New("please specify an output path"), a.window)
		return
	}

	// Load the stego image
	stegoImage, err := loadImage(a.stegoImagePath.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load stego image: %w", err), a.window)
		return
	}

	// Create steganography config
	algorithm := a.algorithm.Selected
	config := stego.Config{
		EmbeddingRate: a.embeddingRate.Value,
	}

	// Set fractal parameters if needed
	if algorithm == AlgorithmFractal {
		// Parse fractal parameters
		iterations, err := strconv.Atoi(a.fractalIterations.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid iterations value: %w", err), a.window)
			return
		}

		threshold, err := strconv.ParseFloat(a.fractalThreshold.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid threshold value: %w", err), a.window)
			return
		}

		config.FractalParams = &stego.FractalParams{
			Type:       a.fractalType.Selected,
			Iterations: iterations,
			Threshold:  threshold,
		}
	}

	// Get the appropriate steganography algorithm
	stegoAlgorithm, err := stego.Factory(algorithm)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Extract the data
	data, err := stegoAlgorithm.Extract(stegoImage, config)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to extract data: %w", err), a.window)
		return
	}

	// Save the extracted data
	err = os.WriteFile(a.outputPath.Text, data, 0644)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save extracted data: %w", err), a.window)
		return
	}

	dialog.ShowInformation("Успех", "Данные успешно извлечены", a.window)
}

func (a *StegoApp) calculateMetrics() {
	originalPath := a.originalImagePath.Text
	stegoPath := a.stegoImagePathForMetrics.Text

	if originalPath == "" || stegoPath == "" {
		dialog.ShowError(errors.New("please select both original and stego images"), a.window)
		return
	}

	originalImg, err := loadImage(originalPath)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	stegoImg, err := loadImage(stegoPath)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	metrics := calculateImageMetrics(originalImg, stegoImg)

	// Display metrics
	metricsText := ""
	for name, value := range metrics {
		metricsText += fmt.Sprintf("%s: %.4f\n", name, value)
	}
	a.metricsText.SetText(metricsText)
}

// Helper functions

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func saveImage(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return png.Encode(file, img)
}

func calculateImageMetrics(original, stego image.Image) map[string]float64 {
	metrics := make(map[string]float64)

	bounds := original.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()

	// MSE and PSNR
	var mseR, mseG, mseB float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			orig := color.RGBAModel.Convert(original.At(x, y)).(color.RGBA)
			steg := color.RGBAModel.Convert(stego.At(x, y)).(color.RGBA)

			diffR := float64(orig.R) - float64(steg.R)
			diffG := float64(orig.G) - float64(steg.G)
			diffB := float64(orig.B) - float64(steg.B)

			mseR += diffR * diffR
			mseG += diffG * diffG
			mseB += diffB * diffB
		}
	}

	mseR /= float64(totalPixels)
	mseG /= float64(totalPixels)
	mseB /= float64(totalPixels)
	mseTotal := (mseR + mseG + mseB) / 3

	metrics["MSE"] = mseTotal
	if mseTotal == 0 {
		metrics["PSNR"] = math.Inf(1)
	} else {
		metrics["PSNR"] = 20 * math.Log10(255/math.Sqrt(mseTotal))
	}

	// Correlation
	var sumOrigR, sumOrigG, sumOrigB float64
	var sumStegR, sumStegG, sumStegB float64
	var sumOrigSqR, sumOrigSqG, sumOrigSqB float64
	var sumStegSqR, sumStegSqG, sumStegSqB float64
	var sumProdR, sumProdG, sumProdB float64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			orig := color.RGBAModel.Convert(original.At(x, y)).(color.RGBA)
			steg := color.RGBAModel.Convert(stego.At(x, y)).(color.RGBA)

			or := float64(orig.R)
			og := float64(orig.G)
			ob := float64(orig.B)
			sr := float64(steg.R)
			sg := float64(steg.G)
			sb := float64(steg.B)

			sumOrigR += or
			sumOrigG += og
			sumOrigB += ob
			sumStegR += sr
			sumStegG += sg
			sumStegB += sb

			sumOrigSqR += or * or
			sumOrigSqG += og * og
			sumOrigSqB += ob * ob
			sumStegSqR += sr * sr
			sumStegSqG += sg * sg
			sumStegSqB += sb * sb

			sumProdR += or * sr
			sumProdG += og * sg
			sumProdB += ob * sb
		}
	}

	n := float64(totalPixels)
	corrR := (n*sumProdR - sumOrigR*sumStegR) /
		math.Sqrt((n*sumOrigSqR-sumOrigR*sumOrigR)*(n*sumStegSqR-sumStegR*sumStegR))
	corrG := (n*sumProdG - sumOrigG*sumStegG) /
		math.Sqrt((n*sumOrigSqG-sumOrigG*sumOrigG)*(n*sumStegSqG-sumStegG*sumStegG))
	corrB := (n*sumProdB - sumOrigB*sumStegB) /
		math.Sqrt((n*sumOrigSqB-sumOrigB*sumOrigB)*(n*sumStegSqB-sumStegB*sumStegB))

	metrics["Correlation (R)"] = corrR
	metrics["Correlation (G)"] = corrG
	metrics["Correlation (B)"] = corrB
	metrics["Correlation (Avg)"] = (corrR + corrG + corrB) / 3

	return metrics
}

func main() {
	stegoApp := NewStegoApp()
	stegoApp.window.ShowAndRun()
}

// weightedLayout is a custom layout that divides space with custom weights
type weightedLayout struct {
	leftWeight  float32
	rightWeight float32
}

func (w *weightedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.NewSize(0, 0)
	}
	leftMin := objects[0].MinSize()
	rightMin := objects[1].MinSize()
	return fyne.NewSize(leftMin.Width+rightMin.Width, fyne.Max(leftMin.Height, rightMin.Height))
}

func (w *weightedLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	if len(objects) != 2 {
		return
	}

	totalWeight := w.leftWeight + w.rightWeight
	leftWidth := float32(containerSize.Width) * (w.leftWeight / totalWeight)

	// Left object
	objects[0].Resize(fyne.NewSize(leftWidth, containerSize.Height))
	objects[0].Move(fyne.NewPos(0, 0))

	// Right object
	rightWidth := containerSize.Width - leftWidth
	objects[1].Resize(fyne.NewSize(rightWidth, containerSize.Height))
	objects[1].Move(fyne.NewPos(leftWidth, 0))
}
