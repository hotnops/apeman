import { HStack, Switch, useColorMode } from "@chakra-ui/react";
import { PiMoonStarsBold, PiSunBold } from "react-icons/pi";

const ColorModeSwitch = () => {
  const { toggleColorMode, colorMode } = useColorMode();
  return (
    <HStack>
      <Switch
        colorScheme="yellow"
        isChecked={colorMode === "dark"}
        onChange={toggleColorMode}
      ></Switch>
      {colorMode === "dark" ? <PiMoonStarsBold /> : <PiSunBold />}
    </HStack>
  );
};

export default ColorModeSwitch;
