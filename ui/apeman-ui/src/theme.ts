import { extendTheme, ThemeConfig } from "@chakra-ui/react";

const config: ThemeConfig = {
  initialColorMode: "light",
};

const theme = extendTheme({
  config,
  colors: {
    gray: {
      "50": "#F4F3F1",
      "100": "#E0DCD7",
      "200": "#CCC6BD",
      "300": "#B7AFA3",
      "400": "#A3998A",
      "500": "#8F8370",
      "600": "#72685A",
      "700": "#564E43",
      "800": "#39342D",
      "900": "#1D1A16",
    },
  },
});

export default theme;
