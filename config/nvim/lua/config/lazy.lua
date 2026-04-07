-- Minimal neovim config (LazyVim removed - use helix/hx instead)
-- This config is intentionally minimal since helix is the primary editor.
-- If you need neovim, it works without the LazyVim overhead.

-- Basic settings
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.tabstop = 4
vim.opt.shiftwidth = 4
vim.opt.expandtab = true
vim.opt.termguicolors = true

-- Leader key
vim.g.mapleader = " "
vim.g.maplocalleader = "\\"

-- Disable netrw (file explorer not needed, lf handles this)
vim.g.loaded_netrwPlugin = 1
