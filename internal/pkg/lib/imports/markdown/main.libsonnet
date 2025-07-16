local elem(name, attrOrChildren=[], childrenOrNull=null) =
  local actualAttr = if childrenOrNull != null then attrOrChildren else null;
  local actualChildren = if childrenOrNull != null then childrenOrNull else attrOrChildren;
  local arrayChildren = if std.type(actualChildren) == 'array' then actualChildren else [actualChildren];
  std.prune([name, actualAttr] + arrayChildren);

{
  Document(children=[]): elem('Document', children),
  FencedCodeBlock(code, language=null, blankPreviousLines=true): elem('FencedCodeBlock', { language: language, blankPreviousLines: blankPreviousLines }, code),
  Emphasis(level, children): elem('Emphasis', { level: level }, children),
  Em(children): self.Emphasis(1, children),
  Strong(children): self.Emphasis(2, children),
  Heading(level, children, blankPreviousLines=true): elem('Heading', { level: level, blankPreviousLines: blankPreviousLines }, children),
  Heading1(children, blankPreviousLines=true): self.Heading(1, children, blankPreviousLines),
  Heading2(children, blankPreviousLines=true): self.Heading(2, children, blankPreviousLines),
  Heading3(children, blankPreviousLines=true): self.Heading(3, children, blankPreviousLines),
  Heading4(children, blankPreviousLines=true): self.Heading(4, children, blankPreviousLines),
  Heading5(children, blankPreviousLines=true): self.Heading(5, children, blankPreviousLines),
  Heading6(children, blankPreviousLines=true): self.Heading(6, children, blankPreviousLines),
  Link(text, destination): elem('Link', { destination: destination }, text),
  Paragraph(children=[], blankPreviousLines=true): elem('Paragraph', { blankPreviousLines: blankPreviousLines }, children),
}
