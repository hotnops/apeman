import { ReactNode, useState } from "react";

interface Props {
  children: ReactNode;
  iconColor: string;
  hoverColor: string;
}

const HoverIcon = ({ children, iconColor, hoverColor }: Props) => {
  const [color, setColor] = useState(iconColor);

  const style = { color, cursor: "pointer" };

  return (
    <div
      style={style}
      onMouseEnter={() => setColor(hoverColor)}
      onMouseLeave={() => setColor(iconColor)}
    >
      {children}
    </div>
  );
};

export default HoverIcon;
