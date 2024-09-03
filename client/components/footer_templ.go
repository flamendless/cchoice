// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.707
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "cchoice/client/components/svg"

func Footer() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<footer class=\"p-2 bg-cchoicesoft dark:bg-cchoicesoft\"><div class=\"inline-flex items-center justify-center w-full mb-1\"><hr class=\"w-8/12 h-px my-8 bg-cchoice border-0 dark:bg-cchoice\"><span class=\"absolute px-3 font-medium text-cchoice -translate-x-1/2 bg-cchoicesoft left-1/2 dark:text-cchoice dark:bg-cchoicesoft\">MORE</span></div><div class=\"mx-auto mt-2 max-w-screen-xl\"><div class=\"p-4\"><div class=\"grid grid-cols-4 gap-4 justify-evenly\"><div><h2 class=\"mb-6 text-sm font-semibold text-black-900 uppercase dark:text-black\">GET IN TOUCH</h2><ul class=\"text-black-600 dark:text-black-400\"><li class=\"mb-2\"><a href=\"https://maps.app.goo.gl/JZCZfbseZuh7eYZg7\" class=\"hover:underline decoration-cchoice flex items-center\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = svg.Map().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p class=\"ml-4 max-w-24\">General Trias, Cavite, 4107</p></a></li><li class=\"mb-2\"><a href=\"mailto:cchoicesales23@gmail.com\" class=\"hover:underline decoration-cchoice flex items-center\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = svg.Mail().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p class=\"ml-4 max-w-24\">cchoicesales23@gmail.com</p></a></li><li class=\"mb-2\"><a href=\"viber://chat?number=09976894824\" class=\"hover:underline decoration-cchoice flex items-center\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = svg.Phone().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p class=\"ml-4 max-w-24\">09976894824 (Viber)</p></a></li></ul></div><div><h2 class=\"mb-6 text-sm font-semibold text-black-900 uppercase dark:text-black\">Follow us</h2><ul class=\"text-black-600 dark:text-black-400\"><li class=\"mb-4\"><a href=\"https://fb.com\" class=\"hover:underline decoration-cchoice \">Facebook</a></li><li class=\"mb-4\"><a href=\"https://fb.com\" class=\"hover:underline decoration-cchoice \">Lazada</a></li><li class=\"mb-4\"><a href=\"https://fb.com\" class=\"hover:underline decoration-cchoice \">Shopee</a></li><li class=\"mb-4\"><a href=\"\" class=\"hover:underline decoration-cchoice \">TikTok</a></li></ul></div><div><h2 class=\"mb-6 text-sm font-semibold text-black-900 uppercase dark:text-black\">Legal</h2><ul class=\"text-black-600 dark:text-black-400\"><li class=\"mb-4\"><a href=\"/privacy\" class=\"hover:underline decoration-cchoice\">Privacy Policy</a></li><li><a href=\"/terms-and-conditions\" class=\"hover:underline decoration-cchoice\">Terms &amp; Conditions</a></li></ul></div><div><h2 class=\"mb-6 text-sm font-semibold text-black-900 uppercase dark:text-black\">Certification</h2><ul class=\"text-black-600 dark:text-black-400\"><li class=\"mb-4\"><a href=\"/privacy\" class=\"hover:underline decoration-cchoice\">Privacy Policy</a></li></ul></div></div></div><hr class=\"w-8/12 h-0.5 mx-auto my-2 bg-cchoice border-0 rounded md:my-10 dark:bg-cchoice\"><div class=\"sm:flex sm:items-center sm:justify-between mb-4\"><span class=\"text-sm text-black-500 sm:text-center dark:text-black-400\">© 2024 <a href=\"/home\" class=\"text-cchoice hover:underline decoration-cchoice\">C-CHOICE™</a>. All Rights Reserved.</span><div class=\"flex mt-4 space-x-6 sm:justify-center sm:mt-0\"><a href=\"https://www.tiktok.com/@cchoicesales?_t=8pPsHyIgtF4&amp;_r=1\" class=\"text-black-500 hover:text-cchoice dark:hover:text-cchoice\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = svg.TikTok().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a> <a href=\"#\" class=\"text-black-500 hover:text-cchoice dark:hover:text-cchoice\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = svg.FB().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a> <a href=\"#\" class=\"text-black-500 hover:text-cchoice dark:hover:text-cchoice\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = svg.IG().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a></div></div></div></footer>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
