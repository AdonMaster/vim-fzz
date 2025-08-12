"
" Fzz Plugin
"

" Global
command! -nargs=0 Fzz call Fzz()
"
function! FzzClose()
    call s:Close()
endfunction



" //
let s:instance = -1
let s:q = ''
let s:data = []
let s:server = -1


" // 
function! Fzz()

    " Is server running
    if type(s:server) != v:t_job

        let l:dir = fnamemodify(resolve(expand('<sfile>:p')), ':h')
        let l:path = l:dir . '/bin/fzz.exe'

        " Job start
        let s:server = job_start([l:path, getcwd()], {
                    \ 'out_cb': funcref('s:ServerCb'), 
                    \ 'in_io': 'pipe', 
                    \ 'out_io': 'pipe', 
                    \ 'err_io': 'null'
                    \ })

        " Check again
        if type(s:server) != v:t_job
            echo 'Error: Failed to start fzz'
            return
        endif
    endif


    "// Initialization
    let s:data = []
    let s:q = ''
       

    "//
    let s:instance = popup_create(s:data, {
        \ 'title': ' Fzz - Adon ',
        \ 'pos': 'center',
        \ 'border': [],
        \ 'padding': [],
        \ 'minwidth': 130, 
        \ 'mapping': 0,
        \ 'filter': funcref('s:Filter'),
        \ 'highlight': 'Normal',
    \ })

    "//
    call s:ServerSend()

endfunction



" //
function! s:ServerCb(channel, message)

    let s:data += [a:message]
    if stridx(a:message, '=======>') == 0
        let s:data = []
    endif

    call s:Refresh()
endfunction



"
function! s:Filter(instance_id, key)
    
    " Escape
    if a:key == "\<Esc>"
        call popup_close(a:instance_id)
        return 1
    endif

    " Backspace
    if a:key == "\<BS>" 
        if len(s:q) > 0
            let s:q = s:q[:-2]
        endif

    " Ctrl + Backspace
    elseif a:key == "\<C-BS>" || a:key == "\<C-h>" "
        let s:q = substitute(s:q, '\s*\S\+$', '', '')

    " Valid characters
    elseif len(a:key) == 1 && char2nr(a:key) >= 32 && char2nr(a:key) <= 126 
        let s:q .= a:key

    " //
    else
        return 1
    endif


    " //
    call s:Refresh()

    " // Server request
    call s:ServerSend()

    "
    return 1

endfunction


" //
function! s:ServerSend()
    let l:buff = s:q . "\n"
    let l:result = ch_sendraw(s:server, l:buff)
    echo l:result
endfunction


" //
function! s:Refresh()

    " Title
    let l:title = (len(s:q) > 0) ? ' "' . s:q . '" ' : ' Fzz - Adon '
    call popup_setoptions(s:instance, {'title': l:title})

    " Set data and forces refresh
    call popup_settext(s:instance, s:data)

endfunction


" //
function! s:Close()
    if s:instance != -1
        call popup_close(s:instance)
        let s:instance = -1
    endif
    if type(s:server) == v:t_job
        call job_stop(s:server)
        let s:server = -1
    endif
endfunction
