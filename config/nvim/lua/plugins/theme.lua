return {
	{
		"olimorris/onedarkpro.nvim",
		lazy = false,
		priority = 1000,
		opts = {
			style = "dark", -- dark, darker, cool, deep, warm, warmer
			transparent = true, -- Enable this to disable setting the background color
			terminal_colors = true, -- Configure the colors used when opening a `:terminal` in Neovim
		},
	},
	{
		"LazyVim/LazyVim",
		opts = {
			colorscheme = "onedark",
		},
	},
}
