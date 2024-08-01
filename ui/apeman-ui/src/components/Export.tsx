import {
  Button,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverCloseButton,
  PopoverContent,
  PopoverHeader,
  PopoverTrigger,
  useToast,
} from "@chakra-ui/react";
import { MdContentCopy } from "react-icons/md";

interface Props {
  object: any;
}

const Export = ({ object }: Props) => {
  const toast = useToast();
  const handleClick = () => {
    navigator.clipboard.writeText(JSON.stringify(object, null, 2));

    toast({
      title: "Copied to clipboard",
      status: "success",
      duration: 2000,
      isClosable: true,
    });
  };
  return (
    <Button onClick={handleClick}>
      <MdContentCopy></MdContentCopy>
    </Button>
  );
};

export default Export;
