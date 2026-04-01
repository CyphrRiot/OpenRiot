-- Avante.nvim - Zed-like AI Assistant for Neovim
-- OpenRouter.ai backend configuration

return {
  {
    "yetone/avante.nvim",
    build = "make",
    event = "VeryLazy",
    version = false,
    opts = {
      -- Core settings
      provider = "openrouter",
      auto_suggestions_provider = "openrouter",
      instructions_file = "avante.md",

      -- OpenRouter configuration
      providers = {
        openrouter = {
          __inherited_from = "openai",
          endpoint = "https://openrouter.ai/api/v1",
          api_key_name = "OPENROUTER_API_KEY",
          model = "minimax/minimax-m2.7",
          timeout = 60000,
          extra_request_body = {
            temperature = 0.7,
            max_tokens = 131072,
          },
        },
      },

      -- UI & behaviour
      behaviour = {
        auto_suggestions = true,
        auto_apply_diff = true,
        support_paste_from_clipboard = true,
      },
    },
    dependencies = {
      "nvim-lua/plenary.nvim",
      "MunifTanjim/nui.nvim",
      "nvim-tree/nvim-web-devicons",
      "MeanderingProgrammer/render-markdown.nvim",
      "HakonHarnes/img-clip.nvim",
      "hrsh7th/nvim-cmp",
      "folke/snacks.nvim",
      "stevearc/dressing.nvim",
    },
  },
}
