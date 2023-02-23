if stridx(expand('%:p'), "clients/frenyard") != -1
  let &makeprg = 'go build -v && FRENYARD_SCALE=1 ./rethink'
endif
