"
" Fzz Plugin
"

" Global
command! -nargs=0 Fzz call Fzz()
nnoremap <C-p> :Fzz<CR>
"
function! FzzClose()
    call s:Close()
endfunction



" //
let s:is_win = has('win32') || has('win64')
let s:instance = -1
let s:q = ''
let s:data = []
let s:server = -1
let s:debounce_refresh = -1
let s:debounce_refresh_delay = 22
let s:loading = ''
let s:selected = 0



" // 
function! Fzz()

    " Is server running
    if type(s:server) != v:t_job

        let l:dir = fnamemodify(resolve(expand('<sfile>:p')), ':h')
        let l:path = l:dir . '/bin/fzz' . (s:is_win ? '.exe' : '')

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
    call ResetData()
    let s:q = ''
       

    "//
    let s:instance = popup_create(s:data, {
        \ 'title': ' Fzz - Adon ',
        \ 'pos': 'center',
        \ 'border': [],
        \ 'padding': [],
        \ 'minwidth': 40, 
        \ 'mapping': 0,
        \ 'filter': funcref('s:Filter'),
        \ 'highlight': 'Normal',
    \ })

    "//
    call s:ServerSend()

endfunction



" //
function! s:ServerCb(channel, message)

    if a:message == '<bof>'
        let s:selected = ResetData()
        let s:loading = ' Loading... '
    elseif a:message == '<eof>'
        let s:loading = ''
    elseif match(a:message, '^<debug ') == 0
        " do nothing as well
    else 
        let s:data += [a:message]
    endif

    " loading title
    call popup_setoptions(s:instance, {'title': s:loading})


    " little debounce logic in here
    if s:debounce_refresh != -1
        call timer_stop(s:debounce_refresh)
    endif
    let s:debounce_refresh = timer_start(s:debounce_refresh_delay, {-> s:Refresh()})

endfunction



"
function! s:Filter(instance_id, key)
    
    " Escape or <C-p>
    if a:key == "\<Esc>" || a:key == "\<C-p>"
        call popup_close(a:instance_id)
        return 1

    endif
    " Enter
    if a:key == "\<Enter>"
        call s:OpenFile()
        return 1
    endif

    " flags
    let l:should_refresh = v:true
    let l:should_send = v:true

    " Backspace
    if a:key == "\<BS>" 
        if len(s:q) > 0
            let s:q = s:q[:-2]
        endif

    " Ctrl + Backspace
    elseif a:key == "\<C-BS>" || a:key == "\<C-h>" "
        let s:q = substitute(s:q, '\s*\S\+$', '', '')

    " ctrl + j
    elseif a:key == "\<C-j>"
        let s:selected += 1
        let s:selected = s:selected % len(s:data)
        let l:should_send = v:false

    " ctrl + k
    elseif a:key == "\<C-k>"
        let s:selected -= 1
        let s:selected = (s:selected + len(s:data)) % len(s:data)
        let l:should_send = v:false

    " Valid characters
    elseif len(a:key) == 1 && char2nr(a:key) >= 32 && char2nr(a:key) <= 126 
        let s:q .= a:key

    " //
    else
        return 1
    endif


    " //
    if l:should_refresh
        call s:Refresh()
    endif

    " // Server request
    if l:should_send
        call s:ServerSend()
    endif

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

    "
    let l:n = len(s:data)


    " Title
    let l:title = (s:q != '') ? (' "' . s:q . '" ') : ' Fzz - Adon '
    call popup_setoptions(s:instance, {'title': l:title})

    " //
    let l:display_data = []
    for i in range(l:n)
        let l:s = s:data[i]
        let l:is_dir = l:s =~ '/$'

        let l:prefix = l:is_dir ? '[] ' : "   "
        if i == s:selected
            let l:symbol = l:is_dir ? '[>' : '->'
            let l:prefix = l:symbol . " "
        endif

        let l:display_data += [l:prefix . l:s]
    endfor

    " refresh
    call popup_settext(s:instance, l:display_data)

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


" //
function! s:OpenFile()

    " selected data to form a valid uri
    let l:uri = get(s:data, s:selected, '')
    if l:uri == ''
        return
    endif

    " full path
    let l:path = getcwd() . '/' . l:uri
    
    " is path directory? append to s:q
    if isdirectory(l:path)

        let s:q = l:uri

        "
        call s:Refresh()
        call s:ServerSend()
        
        return
    endif


    " extract the correct string for target file
    let l:target_file = ''
    if filereadable(l:uri)
        let l:target_file = l:uri
    elseif filereadable(l:path)
        let l:target_file = l:path
    else
        echohl ErrorMsg | echom l:path . ' - not a valid file' | echohl None
        return
    endif

    " ------------------------------------
    " close popup and open file
    call s:Close()

    " open a new buffer, or select an existing one
    let l:bufnum = bufnr(l:target_file)
    if l:bufnum > 0
        execute 'buffer' l:bufnum
    else
        execute 'edit' l:path
    endif

endfunction

" //
function! GetBufferList()
    let l:paths = []
    for b in getbufinfo({'buflisted': 1})
        if !empty(b.name)
            call add(l:paths, b.name)
        endif
    endfor
    return l:paths
endfunction


" //
function! ResetData()
    let s:data = GetBufferList()
    if len(s:data) > 0 
        let s:data += ['-------------------------']
    endif
    return len(s:data)
endfunction
